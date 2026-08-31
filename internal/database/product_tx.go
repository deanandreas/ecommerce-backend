package database

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var Ext = []string{".png", ".jpg"}

func (p *pSQL) InsertProductTx(ctx context.Context, r *http.Request, arg ProductData) (*db.GetProductByUserIDRow, error) {
	tx, err := p.Begin(ctx)
	if err != nil {
		return nil, err
	}

	var imageURLs []string

	defer RollBackWithFile(tx, ctx, imageURLs)

	imageURLs, err = SaveUploads(r, Ext, "products/image", "image")
	if err != nil {
		return nil, err
	}

	q := p.WithTx(tx)

	source := arg.Category.Slug
	if source == "" {
		source = arg.Category.Name
	}
	s := strings.ToLower(strings.TrimSpace(source))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	arg.Category.Slug = strings.Trim(s, "-")

	categoryID, err := q.InsertCategory(ctx, arg.Category)
	if err != nil {
		return nil, err
	}
	arg.CategoryID = categoryID

	productID, err := q.InsertProduct(ctx, arg.InsertProductParams)
	if err != nil {
		return nil, err
	}

	iArg := db.InsertProductImageParams{
		ProductID: productID,
	}
	for i, imageURL := range imageURLs {
		if imageURLs[i] == imageURLs[0] {
			iArg.IsDefault = true
		}
		iArg.ImageUrl = imageURL
		if err := q.InsertProductImage(ctx, iArg); err != nil {
			return nil, err
		}
	}

	pArg := db.GetProductByUserIDParams{ID: productID, UserID: arg.UserID}
	product, err := q.GetProductByUserID(ctx, pArg)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (p *pSQL) UpdateProductTx(ctx context.Context, arg UpdateProductData) (*db.GetProductByUserIDRow, error) {
	tx, err := p.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer RollBack(tx, ctx)

	q := p.WithTx(tx)

	slug := arg.Category.Slug
	var categoryID string
	if arg.Category.Name != "" || slug != "" {
		if slug == "" {
			slug = arg.Category.Name
		}
		s := strings.ToLower(slug)
		re := regexp.MustCompile("[^a-z0-9]+")
		s = re.ReplaceAllString(s, "-")
		arg.Category.Slug = strings.Trim(s, "-")

		categoryID, err = q.InsertCategory(ctx, arg.Category)
		if err != nil {
			return nil, err
		}
		var categoryUUID pgtype.UUID
		err = categoryUUID.Scan(categoryID)
		if err != nil {
			return nil, err
		}
		arg.CategoryID = categoryUUID
	}

	err = q.UpdateProduct(ctx, arg.UpdateProductParams)
	if err != nil {
		return nil, err
	}

	pArg := db.GetProductByUserIDParams{ID: arg.ID, UserID: arg.UserID}
	product, err := q.GetProductByUserID(ctx, pArg)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &product, nil
}

func (p *pSQL) UpdateDefaultImageTx(ctx context.Context, arg db.GetProductByUserIDParams, id string) (*db.GetProductByUserIDRow, error) {
	tx, err := p.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer RollBack(tx, ctx)

	q := p.WithTx(tx)

	imageID, err := q.GetDefualtImageID(ctx, arg.ID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		imageID = "00000000-0000-0000-0000-000000000000"
	}

	uArg := db.UpdateDefaultImageParams{IsDefault: false, ID: imageID}
	_, err = q.UpdateDefaultImage(ctx, uArg)
	if err != nil {
		return nil, err
	}

	uArg.ID = id
	uArg.IsDefault = true
	row, err := q.UpdateDefaultImage(ctx, uArg)
	if err != nil {
		return nil, err
	}
	if row == 0 {
		return nil, ErrNotUpdated
	}

	product, err := q.GetProductByUserID(ctx, arg)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &product, nil
}

func (p *pSQL) InsertProductImagesTx(ctx context.Context, r *http.Request, arg db.GetProductByUserIDParams) (*db.GetProductByUserIDRow, error) {
	tx, err := p.Begin(ctx)
	if err != nil {
		return nil, err
	}
	var imageURLs []string

	defer RollBackWithFile(tx, ctx, imageURLs)

	imageURLs, err = SaveUploads(r, Ext, "products/image", "image")
	if err != nil {
		return nil, err
	}

	q := p.WithTx(tx)

	iArg := db.InsertProductImageParams{ProductID: arg.ID}
	for _, imageURL := range imageURLs {
		iArg.ImageUrl = imageURL
		err := q.InsertProductImage(ctx, iArg)
		if err != nil {
			return nil, err
		}
	}

	product, err := q.GetProductByUserID(ctx, arg)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &product, nil
}

func (p *pSQL) DeleteProductImageTx(ctx context.Context, arg db.GetProductImagesParams, userID string) error {
	tx, err := p.Begin(ctx)
	if err != nil {
		return err
	}

	defer RollBack(tx, ctx)

	q := p.WithTx(tx)

	row, err := q.GetProductImages(ctx, arg)
	if err != nil {
		return err
	} else if row.IsDefault {
		return ErrNotDeleted
	}

	if err := DeleteUploads(row.ImageUrl); err != nil {
		slog.Error("failed to delete the file", "error", err)
	}

	iArg := db.DeleteProductImageParams{ID: arg.ID, UserID: userID}
	err = q.DeleteProductImage(ctx, iArg)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
