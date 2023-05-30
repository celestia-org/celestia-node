package blob

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/cosmos/cosmos-sdk/types"
	logging "github.com/ipfs/go-log/v2"

	"github.com/celestiaorg/celestia-app/pkg/shares"
	apptypes "github.com/celestiaorg/celestia-app/x/blob/types"
	"github.com/celestiaorg/nmt/namespace"

	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/share"
	"github.com/celestiaorg/celestia-node/state"
)

var (
	ErrBlobNotFound = errors.New("blob: not found")
	ErrInvalidProof = errors.New("blob: invalid proof")

	log = logging.Logger("blob")
)

// TODO(@vgonkivs): remove after bumping celestia-app
// defaultMinGasPrice is the default min gas price.
const defaultMinGasPrice = 0.001

type Service struct {
	// accessor dials the given celestia-core endpoint to submit blobs.
	accessor *state.CoreAccessor
	// shareGetter retrieves the EDS to fetch all shares from the requested header.
	shareGetter share.Getter
	//  headerGetter fetches header by the provided height
	headerGetter func(context.Context, uint64) (*header.ExtendedHeader, error)
}

func NewService(
	state *state.CoreAccessor,
	getter share.Getter,
	headerGetter func(context.Context, uint64) (*header.ExtendedHeader, error),
) *Service {
	return &Service{
		accessor:     state,
		shareGetter:  getter,
		headerGetter: headerGetter,
	}
}

// Submit sends PFB transaction and reports the height in which it was included.
// Allows sending multiple Blobs atomically synchronously.
// Uses default wallet registered on the Node.
func (s *Service) Submit(ctx context.Context, blobs ...*Blob) (uint64, error) {
	log.Debugw("submitting blobs", "amount", len(blobs))

	var (
		gasLimit = estimateGas(blobs...)
		fee      = int64(defaultMinGasPrice * float64(gasLimit))
		b        = make([]*apptypes.Blob, len(blobs))
	)

	for i, blob := range blobs {
		b[i] = &blob.Blob
	}

	resp, err := s.accessor.SubmitPayForBlob(ctx, types.NewInt(fee), gasLimit, b...)
	if err != nil {
		return 0, err
	}
	return uint64(resp.Height), nil
}

// Get retrieves all the blobs for given namespaces at the given height by commitment.
func (s *Service) Get(ctx context.Context, height uint64, nID namespace.ID, commitment Commitment) (*Blob, error) {
	blob, _, err := s.getByCommitment(ctx, height, nID, commitment)
	if err != nil {
		return nil, err
	}
	return blob, nil
}

// GetProof retrieves all blobs in the given namespaces at the given height by commitment
// and returns their Proof.
func (s *Service) GetProof(
	ctx context.Context,
	height uint64,
	nID namespace.ID,
	commitment Commitment,
) (*Proof, error) {
	_, proof, err := s.getByCommitment(ctx, height, nID, commitment)
	if err != nil {
		return nil, err
	}
	return proof, nil
}

// GetAll returns all blobs under the given namespaces at the given height.
// GetAll can return blobs and an error in case if some requests failed.
func (s *Service) GetAll(ctx context.Context, height uint64, nIDs ...namespace.ID) ([]*Blob, error) {
	header, err := s.headerGetter(ctx, height)
	if err != nil {
		return nil, err
	}

	var (
		resultBlobs = make([][]*Blob, len(nIDs))
		resultErr   = make([]error, len(nIDs))
	)

	wg := sync.WaitGroup{}
	for i, nID := range nIDs {
		wg.Add(1)
		go func(i int, nID namespace.ID) {
			defer wg.Done()
			blobs, err := s.getBlobs(ctx, nID, header.DAH)
			if err != nil {
				resultErr[i] = fmt.Errorf("getting blobs for nID(%s): %s", nID.String(), err)
				return
			}
			resultBlobs[i] = blobs
		}(i, nID)
	}
	wg.Wait()

	blobs := make([]*Blob, 0)
	for _, resBlobs := range resultBlobs {
		if len(resBlobs) > 0 {
			blobs = append(blobs, resBlobs...)
		}
	}

	if len(blobs) == 0 {
		resultErr = append(resultErr, ErrBlobNotFound)
	}
	return blobs, errors.Join(resultErr...)
}

// Included verifies that the blob was included in a specific height.
// To ensure that blob was included in a specific height, we need:
// 1. verify the provided commitment by recomputing it;
// 2. verify the provided Proof against subtree roots that were used in 1.;
func (s *Service) Included(
	ctx context.Context,
	height uint64,
	nID namespace.ID,
	proof *Proof,
	com Commitment,
) (bool, error) {
	// In the current implementation, LNs will have to download all shares to recompute the commitment.
	// To achieve 1. we need to modify Proof structure and to store all subtree roots, that were
	// involved in commitment creation and then call `merkle.HashFromByteSlices`(tendermint package).
	// nmt.Proof is verifying share inclusion by recomputing row roots, so, theoretically, we can do
	// the same but using subtree roots. For this case, we need an extra method in nmt.Proof
	// that will perform all reconstructions,
	// but we have to guarantee that all our stored subtree roots will be on the same height(e.g. one
	// level above shares).
	// TODO(@vgonkivs): rework the implementation to perform all verification without network requests.
	_, resProof, err := s.getByCommitment(ctx, height, nID, com)
	switch err {
	case nil:
	case ErrBlobNotFound:
		return false, nil
	default:
		return false, err
	}
	return true, resProof.equal(*proof)
}

// getByCommitment retrieves the DAH row by row, fetching shares and constructing blobs in order to
// compare Commitments. Retrieving is stopped once the requested blob/proof is found.
func (s *Service) getByCommitment(
	ctx context.Context,
	height uint64,
	nID namespace.ID,
	commitment Commitment,
) (*Blob, *Proof, error) {
	log.Infow("requesting blob",
		"height", height,
		"nID", nID.String())

	header, err := s.headerGetter(ctx, height)
	if err != nil {
		return nil, nil, err
	}

	var (
		rawShares = make([]share.Share, 0)
		proofs    = make(Proof, 0)
		amount    int
		blobShare *shares.Share
	)

	namespacedShares, err := s.shareGetter.GetSharesByNamespace(ctx, header.DAH, nID)
	if err != nil {
		if errors.Is(err, share.ErrNamespaceNotFound) ||
			errors.Is(err, share.ErrNotFound) {
			err = ErrBlobNotFound
		}
		return nil, nil, err
	}
	for _, row := range namespacedShares {
		if len(row.Shares) == 0 {
			break
		}

		rawShares = append(rawShares, row.Shares...)
		proofs = append(proofs, row.Proof)

		// reconstruct the `blobShare` from the first rawShare in range
		// in order to get blob's length(first share will contain this info)
		if blobShare == nil {
			for i, sh := range rawShares {
				bShare, err := shares.NewShare(sh)
				if err != nil {
					return nil, nil, err
				}

				// ensure that the first share is not a NamespacePaddingShare
				// these shares are used to satisfy the non-interactive default rules
				// and are not the part of the blob, so should be removed.
				isPadding, err := bShare.IsPadding()
				if err != nil {
					return nil, nil, err
				}
				if isPadding {
					continue
				}

				blobShare = bShare
				// save the length.
				length, err := blobShare.SequenceLen()
				if err != nil {
					return nil, nil, err
				}
				amount = shares.SparseSharesNeeded(length)
				rawShares = rawShares[i:]
				break
			}
		}

		// move to the next row if the blob is incomplete.
		if amount > len(rawShares) {
			continue
		}

		blob, same, err := constructAndVerifyBlob(rawShares[:amount], commitment)
		if err != nil {
			return nil, nil, err
		}
		if same {
			return blob, &proofs, nil
		}

		// drop info of the checked blob
		rawShares = rawShares[amount:]
		if len(rawShares) > 0 {
			// save proof for the last row in case we have rawShares
			proofs = proofs[len(proofs)-1:]
		} else {
			// otherwise clear proofs
			proofs = nil
		}
		blobShare = nil
	}

	if len(rawShares) == 0 {
		return nil, nil, ErrBlobNotFound
	}

	blob, same, err := constructAndVerifyBlob(rawShares, commitment)
	if err != nil {
		return nil, nil, err
	}
	if same {
		return blob, &proofs, nil
	}

	return nil, nil, ErrBlobNotFound
}

// getBlobs retrieves the DAH and fetches all shares from the requested namespace.ID and converts
// them to Blobs.
func (s *Service) getBlobs(ctx context.Context, nID namespace.ID, root *share.Root) ([]*Blob, error) {
	namespacedShares, err := s.shareGetter.GetSharesByNamespace(ctx, root, nID)
	if err != nil {
		return nil, err
	}
	return SharesToBlobs(namespacedShares.Flatten())
}
