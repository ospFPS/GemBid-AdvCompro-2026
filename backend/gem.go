package main
import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// ============================================================
// CREATE GEM REQUEST
// ============================================================

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

// ============================================================
// GEM RESPONSE
// ============================================================

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

// ============================================================
// POST /api/gems
// ============================================================

// CreateGemHandler godoc
// @Summary Create a gemstone
// @Description Create a gemstone owned by the authenticated verified seller.
// @Tags Gems
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateGemRequest true "Gemstone details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 403 {string} string
// @Failure 500 {string} string
// @Router /api/gems [post]
func CreateGemHandler(w http.ResponseWriter, r *http.Request) {

	// seller_id was added by VerifiedSellerMiddleware.
	sellerID, ok := r.Context().Value(sellerIDKey).(int64)

	if !ok {
		http.Error(
			w,
			"Verified seller account required",
			http.StatusForbidden,
		)
		return
	}

	var request CreateGemRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		http.Error(
			w,
			"Invalid JSON",
			http.StatusBadRequest,
		)
		return
	}

	// ========================================================
	// CLEAN INPUT
	// ========================================================

	request.GemName = strings.TrimSpace(request.GemName)
	request.GemType = strings.TrimSpace(request.GemType)
	request.Shape = strings.TrimSpace(request.Shape)
	request.Color = strings.TrimSpace(request.Color)
	request.Clarity = strings.TrimSpace(request.Clarity)
	request.Origin = strings.TrimSpace(request.Origin)
	request.Treatment = strings.TrimSpace(request.Treatment)
	request.Description = strings.TrimSpace(request.Description)

	// ========================================================
	// VALIDATION
	// ========================================================

	if request.GemName == "" {
		http.Error(
			w,
			"gem_name is required",
			http.StatusBadRequest,
		)
		return
	}

	if request.GemType == "" {
		http.Error(
			w,
			"gem_type is required",
			http.StatusBadRequest,
		)
		return
	}

	if request.Carat != nil && *request.Carat <= 0 {
		http.Error(
			w,
			"carat must be greater than 0",
			http.StatusBadRequest,
		)
		return
	}

	if request.Length != nil && *request.Length <= 0 {
		http.Error(
			w,
			"length must be greater than 0",
			http.StatusBadRequest,
		)
		return
	}

	if request.Width != nil && *request.Width <= 0 {
		http.Error(
			w,
			"width must be greater than 0",
			http.StatusBadRequest,
		)
		return
	}

	if request.Height != nil && *request.Height <= 0 {
		http.Error(
			w,
			"height must be greater than 0",
			http.StatusBadRequest,
		)
		return
	}

	// ========================================================
	// INSERT GEM
	// ========================================================

	var gem GemResponse

	err := DB.QueryRow(
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
	).Scan(
		&gem.GemID,
		&gem.SellerID,
		&gem.GemName,
		&gem.GemType,
		&gem.Carat,
		&gem.Length,
		&gem.Width,
		&gem.Height,
		&gem.Shape,
		&gem.Color,
		&gem.Clarity,
		&gem.Origin,
		&gem.Treatment,
		&gem.Description,
		&gem.CreatedAt,
		&gem.UpdatedAt,
	)

	if err != nil {
		http.Error(
			w,
			"Could not create gem",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"message": "Gem created successfully",
			"gem":     gem,
		},
	)
}
