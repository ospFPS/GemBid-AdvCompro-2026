package gembid

import (
	"encoding/json"
	"net/http"
	"time"
)

type CreateGemRequest struct {
	GemName     string   `json:"gem_name"`
	GemType     string   `json:"gem_type"`
	Carat       *float64 `json:"carat"`
	Length      *float64 `json:"length"`
	Width       *float64 `json:"width"`
	Height      *float64 `json:"height"`
	Shape       string   `json:"shape"`
	Color       string   `json:"color"`
	Clarity     string   `json:"clarity"`
	Origin      string   `json:"origin"`
	Treatment   string   `json:"treatment"`
	Description string   `json:"description"`
}

type GemResponse struct {
	GemID       int64     `json:"gem_id"`
	SellerID    int64     `json:"seller_id"`
	GemName     string    `json:"gem_name"`
	GemType     string    `json:"gem_type"`
	Carat       *float64  `json:"carat"`
	Length      *float64  `json:"length"`
	Width       *float64  `json:"width"`
	Height      *float64  `json:"height"`
	Shape       *string   `json:"shape"`
	Color       *string   `json:"color"`
	Clarity     *string   `json:"clarity"`
	Origin      *string   `json:"origin"`
	Treatment   *string   `json:"treatment"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func CreateGemHandler(w http.ResponseWriter, r *http.Request) {

	sellerID, ok := r.Context().Value(sellerIDKey).(int64)

	if !ok {
		http.Error(w, "Verified seller account required", http.StatusForbidden)
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

	var gem GemResponse

	err := scanGem(DB.QueryRow(
		r.Context(),
		`
		INSERT INTO public.gems (
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
			description
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			NULLIF($8, ''),
			NULLIF($9, ''),
			NULLIF($10, ''),
			NULLIF($11, ''),
			NULLIF($12, ''),
			NULLIF($13, '')
		)
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
		sellerID,
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
	), &gem)

	if err != nil {
		http.Error(w, "Could not create gem", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"message": "Gem created successfully",
			"gem":     gem,
		},
	)
}
