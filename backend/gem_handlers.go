package gembid

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

func cleanAndValidateGemRequest(request *CreateGemRequest) string {
	request.GemName = strings.TrimSpace(request.GemName)
	request.GemType = strings.TrimSpace(request.GemType)
	request.Shape = strings.TrimSpace(request.Shape)
	request.Color = strings.TrimSpace(request.Color)
	request.Clarity = strings.TrimSpace(request.Clarity)
	request.Origin = strings.TrimSpace(request.Origin)
	request.Treatment = strings.TrimSpace(request.Treatment)
	request.Description = strings.TrimSpace(request.Description)

	if request.GemName == "" {
		return "gem_name is required"
	}

	if request.GemType == "" {
		return "gem_type is required"
	}
	if len(request.GemName) > 150 || len(request.GemType) > 100 ||
		len(request.Shape) > 100 || len(request.Color) > 100 || len(request.Clarity) > 100 ||
		len(request.Origin) > 150 || len(request.Treatment) > 150 || len(request.Description) > 5000 {
		return "gem field is too long"
	}

	if request.Carat != nil && (*request.Carat <= 0 || *request.Carat > 9999999.999) {
		return "carat must be greater than 0 and within database range"
	}

	if request.Length != nil && (*request.Length <= 0 || *request.Length > 99999999.99) {
		return "length must be greater than 0 and within database range"
	}

	if request.Width != nil && (*request.Width <= 0 || *request.Width > 99999999.99) {
		return "width must be greater than 0 and within database range"
	}

	if request.Height != nil && (*request.Height <= 0 || *request.Height > 99999999.99) {
		return "height must be greater than 0 and within database range"
	}

	return ""
}

func scanGem(row pgx.Row, gem *GemResponse) error {
	return row.Scan(
		&gem.GemID, &gem.SellerID, &gem.GemName, &gem.GemType,
		&gem.Carat, &gem.Length, &gem.Width, &gem.Height,
		&gem.Shape, &gem.Color, &gem.Clarity, &gem.Origin,
		&gem.Treatment, &gem.Description, &gem.CreatedAt, &gem.UpdatedAt,
	)
}

func ListGemsHandler(w http.ResponseWriter, r *http.Request) {

	rows, err := DB.Query(
		r.Context(),
		`
		SELECT
			gem_id,
			seller_id,
			gem_name,
			gem_type,
			carat,
			length,
			width,
			height,
			shape,
			color,
			clarity,
			origin,
			treatment,
			description,
			created_at,
			updated_at
		FROM public.gems
		ORDER BY created_at DESC
		`,
	)

	if err != nil {
		http.Error(w, "Could not fetch gems", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	gems := make([]GemResponse, 0)

	for rows.Next() {

		var gem GemResponse

		err := scanGem(rows, &gem)

		if err != nil {
			http.Error(w, "Could not read gem data", http.StatusInternalServerError)
			return
		}

		gems = append(gems, gem)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Could not finish reading gems", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(gems)
}

func GetGemHandler(w http.ResponseWriter, r *http.Request) {

	gemID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil || gemID <= 0 {
		http.Error(w, "Invalid gem ID", http.StatusBadRequest)
		return
	}

	var gem GemResponse

	err = scanGem(DB.QueryRow(
		r.Context(),
		`
		SELECT
			gem_id,
			seller_id,
			gem_name,
			gem_type,
			carat,
			length,
			width,
			height,
			shape,
			color,
			clarity,
			origin,
			treatment,
			description,
			created_at,
			updated_at
		FROM public.gems
		WHERE gem_id = $1
		`,
		gemID,
	), &gem)

	if err == pgx.ErrNoRows {
		http.Error(w, "Gem not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Could not fetch gem", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(gem)
}

func UpdateGemHandler(w http.ResponseWriter, r *http.Request) {

	sellerID, ok := r.Context().Value(sellerIDKey).(int64)

	if !ok {
		http.Error(w, "Verified seller account required", http.StatusForbidden)
		return
	}

	gemID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil || gemID <= 0 {
		http.Error(w, "Invalid gem ID", http.StatusBadRequest)
		return
	}

	var request CreateGemRequest

	if err := decodeJSON(r, &request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if validationError := cleanAndValidateGemRequest(&request); validationError != "" {
		http.Error(w, validationError, http.StatusBadRequest)
		return
	}

	var ownerID int64
	var locked bool

	err = DB.QueryRow(
		r.Context(),
		`
		SELECT
			g.seller_id,
			EXISTS (
				SELECT 1
				FROM public.listings l
				WHERE l.gem_id = g.gem_id
				AND l.status IN (
					'PUBLISHED',
					'ACTIVE',
					'SOLD',
					'SUSPENDED'
				)
			)
		FROM public.gems g
		WHERE g.gem_id = $1
		`,
		gemID,
	).Scan(
		&ownerID,
		&locked,
	)

	if err == pgx.ErrNoRows {
		http.Error(w, "Gem not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Could not verify gem", http.StatusInternalServerError)
		return
	}

	if ownerID != sellerID {
		http.Error(w, "You do not own this gem", http.StatusForbidden)
		return
	}

	if locked {
		http.Error(w, "Gem cannot be edited while it has a published, active, sold, or suspended listing", http.StatusConflict)
		return
	}

	var gem GemResponse

	err = scanGem(DB.QueryRow(
		r.Context(),
		`
		UPDATE public.gems
		SET
			gem_name = $1,
			gem_type = $2,
			carat = $3,
			length = $4,
			width = $5,
			height = $6,
			shape = NULLIF($7, ''),
			color = NULLIF($8, ''),
			clarity = NULLIF($9, ''),
			origin = NULLIF($10, ''),
			treatment = NULLIF($11, ''),
			description = NULLIF($12, ''),
			updated_at = NOW()
		WHERE gem_id = $13
		AND seller_id = $14
		RETURNING
			gem_id,
			seller_id,
			gem_name,
			gem_type,
			carat,
			length,
			width,
			height,
			shape,
			color,
			clarity,
			origin,
			treatment,
			description,
			created_at,
			updated_at
		`,
		request.GemName,
		request.GemType,
		request.Carat,
		request.Length,
		request.Width,
		request.Height,
		request.Shape,
		request.Color,
		request.Clarity,
		request.Origin,
		request.Treatment,
		request.Description,
		gemID,
		sellerID,
	), &gem)

	if err != nil {
		http.Error(w, "Could not update gem", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"message": "Gem updated successfully",
			"gem":     gem,
		},
	)
}

func DeleteGemHandler(w http.ResponseWriter, r *http.Request) {

	sellerID, ok := r.Context().Value(sellerIDKey).(int64)

	if !ok {
		http.Error(w, "Verified seller account required", http.StatusForbidden)
		return
	}

	gemID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil || gemID <= 0 {
		http.Error(w, "Invalid gem ID", http.StatusBadRequest)
		return
	}

	var ownerID int64

	err = DB.QueryRow(
		r.Context(),
		`
		SELECT seller_id
		FROM public.gems
		WHERE gem_id = $1
		`,
		gemID,
	).Scan(&ownerID)

	if err == pgx.ErrNoRows {
		http.Error(w, "Gem not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Could not verify gem", http.StatusInternalServerError)
		return
	}

	if ownerID != sellerID {
		http.Error(w, "You do not own this gem", http.StatusForbidden)
		return
	}

	_, err = DB.Exec(
		r.Context(),
		`
		DELETE FROM public.gems
		WHERE gem_id = $1
		AND seller_id = $2
		`,
		gemID,
		sellerID,
	)

	if err != nil {

		if isPGError(err, "23503") {

			http.Error(w, "Gem cannot be deleted because it is referenced by a listing or another record", http.StatusConflict)
			return
		}

		http.Error(w, "Could not delete gem", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(
		map[string]string{
			"message": "Gem deleted successfully",
		},
	)
}
