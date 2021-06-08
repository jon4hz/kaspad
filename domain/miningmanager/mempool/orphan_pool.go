package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/model"
	"github.com/pkg/errors"
)

type idToOrphan map[externalapi.DomainTransactionID]*model.OrphanTransaction
type previousOutpointToOrphans map[externalapi.DomainOutpoint]idToOrphan

type orphansPool struct {
	mempool                   *mempool
	allOrphans                idToOrphan
	orphansByPreviousOutpoint previousOutpointToOrphans
	lastExpireScan            uint64
}

func newOrphansPool(mp *mempool) *orphansPool {
	return &orphansPool{
		mempool:                   mp,
		allOrphans:                idToOrphan{},
		orphansByPreviousOutpoint: previousOutpointToOrphans{},
		lastExpireScan:            0,
	}
}

func (op *orphansPool) maybeAddOrphan(transaction *externalapi.DomainTransaction,
	missingParents []*externalapi.DomainTransactionID, isHighPriority bool) error {

	panic("orphansPool.maybeAddOrphan not implemented") // TODO (Mike)
}

func (op *orphansPool) processOrphansAfterAcceptedTransaction(acceptedTransaction *model.MempoolTransaction) (
	acceptedOrphans []*model.MempoolTransaction, err error) {

	panic("orphansPool.processOrphansAfterAcceptedTransaction not implemented") // TODO (Mike)
}

func (op *orphansPool) unorphanTransaction(orphanTransactionID *externalapi.DomainTransactionID) (model.MempoolTransaction, error) {
	panic("orphansPool.unorphanTransaction not implemented") // TODO (Mike)
}

func (op *orphansPool) removeOrphan(orphanTransactionID *externalapi.DomainTransactionID, removeRedeemers bool) error {
	orphanTransaction, ok := op.allOrphans[*orphanTransactionID]
	if !ok {
		return nil
	}

	delete(op.allOrphans, orphanTransactionID)

	for i, input := range orphanTransaction.Transaction.Inputs {
		orphans, ok := op.orphansByPreviousOutpoint[input.PreviousOutpoint]
		if !ok {
			return errors.Errorf("Input No. %d of %s (%s) doesn't exist in orphansByPreviousOutpoint",
				i, orphanTransactionID, input.PreviousOutpoint)
		}
		delete(orphans, *orphanTransactionID)
		if len(orphans) == 0 {
			delete(op.orphansByPreviousOutpoint, input.PreviousOutpoint)
		}
	}

	if removeRedeemers {
		outpoint := externalapi.DomainOutpoint{TransactionID: *orphanTransactionID}
		for i := range orphanTransaction.Transaction.Outputs {
			outpoint.Index = uint32(i)
			for _, orphan := range op.orphansByPreviousOutpoint[outpoint] {
				// Recursive call is bound by size of orphan pool (which is very small)
				err := op.removeOrphan(orphan.TransactionID(), true)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (op *orphansPool) expireOrphanTransactions() error {
	virtualDAAScore, err := op.mempool.virtualDAAScore()
	if err != nil {
		return err
	}

	if virtualDAAScore-op.lastExpireScan < op.mempool.config.orphanExpireScanIntervalDAAScore {
		return nil
	}

	for _, orphanTransaction := range op.allOrphans {
		// Never expire high priority transactions
		if orphanTransaction.IsHighPriority {
			continue
		}

		// Remove all transactions whose addedAtDAAScore is older then transactionExpireIntervalDAAScore
		if virtualDAAScore-orphanTransaction.AddedAtDAAScore > op.mempool.config.orphanExpireIntervalDAAScore {
			err = op.removeOrphan(orphanTransaction.TransactionID())
			if err != nil {
				return err
			}
		}
	}

	op.lastExpireScan = virtualDAAScore
	return nil
}
