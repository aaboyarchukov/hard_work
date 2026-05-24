package inconsistency

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

func (r *Repository) BatchCreateAtomic(ctx context.Context, docs []*domain.UploadDocument) (map[*domain.UploadDocument]string, error) {
	const op = "minio.BatchCreateAtomic"

	if len(docs) == 0 {
		return make(map[*domain.UploadDocument]string), nil
	}

	var (
		results = make(map[*domain.UploadDocument]string, len(docs))
		errs    = make(map[*domain.UploadDocument]error)
		mu      sync.Mutex
		wg      sync.WaitGroup
	)

	for _, doc := range docs {
		wg.Go(func() {
			link, err := r.Create(ctx, repository.StorageFileInput{
				Id:          uuid.New(),
				Name:        doc.Name,
				File:        bytes.NewReader(doc.File),
				Size:        int64(len(doc.File)),
				ContentType: utils.DetectContentType(doc.File),
			})

			mu.Lock()
			defer mu.Unlock()

			results[doc] = link
			if err != nil {
				errs[doc] = err
			}
		})
	}

	wg.Wait()

	if len(errs) > 0 {
		keys := make(map[string]struct{}, len(results))
		for _, link := range results {
			keys[link] = struct{}{}
		}
		r.RollbackUploaded(ctx, keys)

		return nil, fmt.Errorf("%s: %d/%d failed: %v", op, len(errs), len(docs), errs)
	}

	return results, nil
}
