package inconsistency

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

func (r *Repository) UploadAtomicBatch(
	ctx context.Context,
	files map[repository.FileID]repository.StorageFileInput,
) map[repository.S3Link]repository.StorageFileOutput {
	var (
		uploaded = make(map[repository.S3Link]repository.StorageFileOutput, len(files))
	)

	worker, errCtx := errgroup.WithContext(ctx)

	for _, file := range files {
		worker.Go(func() error {
			s3Link, err := r.Create(errCtx, file)
			if err != nil {
				return err
			}
			uploaded[s3Link] = repository.StorageFileOutput{
				FileType: file.FileType,
			}

			return nil
		})
	}

	if err := worker.Wait(); err != nil {
		keys := make(map[string]struct{}, len(uploaded))
		for key := range uploaded {
			keys[key] = struct{}{}
		}
		r.RollbackUploaded(ctx, keys)

		return nil
	}

	return uploaded
}

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
