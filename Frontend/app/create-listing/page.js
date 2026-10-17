
"use client";

import { useState } from "react";
import { api, resizeImage } from "../lib/api";

const AUCTION_DURATIONS = [
    { label: "5 minutes (demo)", minutes: 5 },
    { label: "1 hour", minutes: 60 },
    { label: "1 day", minutes: 60 * 24 },
    { label: "3 days", minutes: 60 * 24 * 3 },
    { label: "7 days", minutes: 60 * 24 * 7 },
];

export default function CreateListing() {
    const [gemstoneName, setGemstoneName] = useState("");
    const [gemstoneType, setGemstoneType] = useState("Sapphire");
    const [description, setDescription] = useState("");
    const [saleMethod, setSaleMethod] = useState("Fixed Price");
    const [price, setPrice] = useState("");
    const [minimumIncrement, setMinimumIncrement] = useState("");
    const [reservePrice, setReservePrice] = useState("");
    const [images, setImages] = useState([]);
    const [certificate, setCertificate] = useState(null);
    const [message, setMessage] = useState("");
    const [durationMinutes, setDurationMinutes] = useState(60 * 24);
    const [busy, setBusy] = useState(false);

    // Gem (+ photos, first one is the main photo) -> DRAFT listing. Certificates are not uploaded yet.
    async function createDraft() {
        const gem = (await api("/api/gems", {
            method: "POST",
            body: { gem_name: gemstoneName, gem_type: gemstoneType, description },
        })).gem;
        for (const file of images) {
            await api(`/api/gems/${gem.gem_id}/images`, {
                method: "POST",
                body: { image_url: await resizeImage(file) },
            });
        }
        return (await api("/api/listings", { method: "POST", body: { gem_id: gem.gem_id } })).listing;
    }

    async function run(task) {
        setBusy(true);
        setMessage("");
        try {
            await task();
        } catch (err) {
            setMessage(err.message);
        }
        setBusy(false);
    }

    function handleSaveDraft() {
        if (!gemstoneName.trim()) {
            setMessage("Listing name is required.");
            return;
        }
        run(async () => {
            const listing = await createDraft();
            setMessage(`Draft saved (listing #${listing.listing_id}).`);
        });
    }

    function handlePublish(event) {
        event.preventDefault();
        run(async () => {
            const listing = await createDraft();
            if (saleMethod === "Fixed Price") {
                await api(`/api/listings/${listing.listing_id}/publish`, {
                    method: "PUT",
                    body: { sale_type: "FIXED_PRICE", fixed_price: Number(price) },
                });
            } else {
                await api(`/api/listings/${listing.listing_id}/publish`, {
                    method: "PUT",
                    body: { sale_type: "AUCTION" },
                });
                const start = new Date();
                await api("/api/auctions", {
                    method: "POST",
                    body: {
                        listing_id: listing.listing_id,
                        starting_price: Number(price),
                        minimum_increment: Number(minimumIncrement),
                        reserve_price: reservePrice ? Number(reservePrice) : null,
                        start_time: start.toISOString(),
                        end_time: new Date(start.getTime() + durationMinutes * 60000).toISOString(),
                    },
                });
            }
            window.location.href = `/item?id=${listing.listing_id}`;
        });
    }

    return (
        <main className="create-listing-page">
            <a href="/my-listings" className="create-listing-back-link">
                ← Back to My Listings
            </a>
            <div className="create-listing-heading">
                <h1>Create Listing</h1>
                <p>Add your gemstone or jewelry item to GemBid.</p>
            </div>
            <form className="create-listing-form" onSubmit={handlePublish}>
                <section className="listing-form-section">
                    <h2>Item Details</h2>
                    <div className="listing-form-group">
                        <label htmlFor="gemstoneName">Listing Name</label>
                        <input
                            id="gemstoneName"
                            type="text"
                            value={gemstoneName}
                            onChange={(event) => setGemstoneName(event.target.value)}
                            placeholder="e.g. Blue Sapphire Ring"
                            required
                        />
                    </div>
                    <div className="listing-form-group">
                        <label htmlFor="gemstoneType">Gemstone Type</label>
                        <select
                            id="gemstoneType"
                            value={gemstoneType}
                            onChange={(event) => setGemstoneType(event.target.value)}
                        >
                            <option value="Sapphire">Sapphire</option>
                            <option value="Ruby">Ruby</option>
                            <option value="Emerald">Emerald</option>
                            <option value="Diamond">Diamond</option>
                            <option value="Other">Other</option>
                        </select>
                    </div>
                    <div className="listing-form-group">
                        <label htmlFor="description">Description</label>
                        <textarea
                            id="description"
                            value={description}
                            onChange={(event) => setDescription(event.target.value)}
                            placeholder="Describe the gemstone, condition, size, and other details."
                            rows="5"
                            required
                        />
                    </div>
                </section>
                <section className="listing-form-section">
                    <h2>Gemstone Images</h2>
                    <div className="listing-form-group">
                        <label htmlFor="images">Upload Images</label>
                        <input
                            id="images"
                            type="file"
                            accept="image/*"
                            multiple
                            onChange={(event) => setImages(Array.from(event.target.files || []))}
                        />
                        {images.length > 0 && (
                            <>
                                <p className="selected-file-info">
                                    {images.length} image(s) selected. The first one is the main photo.
                                </p>
                                <div style={{ display: "flex", gap: "8px", flexWrap: "wrap" }}>
                                    {images.map((file) => (
                                        <img
                                            key={file.name + file.size}
                                            src={URL.createObjectURL(file)}
                                            alt={file.name}
                                            style={{ width: "96px", height: "96px", objectFit: "cover", borderRadius: "8px" }}
                                        />
                                    ))}
                                </div>
                            </>
                        )}
                    </div>
                </section>

                <section className="listing-form-section">
                    <h2>Sale Details</h2>
                    <div className="listing-form-group">
                        <label htmlFor="saleMethod">Sale Method</label>
                        <select
                            id="saleMethod"
                            value={saleMethod}
                            onChange={(event) => {
                                setSaleMethod(event.target.value);
                                setMessage("");
                            }}
                        >
                            <option value="Fixed Price">Fixed Price</option>
                            <option value="Auction">Auction</option>
                        </select>
                    </div>
                    {saleMethod === "Fixed Price" ? (
                        <div className="listing-form-group">
                            <label htmlFor="price">Selling Price (THB)</label>
                            <input
                                id="price"
                                type="number"
                                min="1"
                                step="0.01"
                                value={price}
                                onChange={(event) => setPrice(event.target.value)}
                                placeholder="e.g. 5200"
                                required
                            />
                        </div>
                    ) : (
                        <>
                            <div className="listing-form-group">
                                <label htmlFor="price">Starting Price (THB)</label>
                                <input
                                    id="price"
                                    type="number"
                                    min="1"
                                    step="0.01"
                                    value={price}
                                    onChange={(event) => setPrice(event.target.value)}
                                    placeholder="e.g. 3000"
                                    required
                                />
                            </div>
                            <div className="listing-form-group">
                                <label htmlFor="minimumIncrement">Minimum Bid Increment (THB)</label>
                                <input
                                    id="minimumIncrement"
                                    type="number"
                                    min="0.01"
                                    step="0.01"
                                    value={minimumIncrement}
                                    onChange={(event) => setMinimumIncrement(event.target.value)}
                                    placeholder="e.g. 100"
                                    required
                                />
                            </div>
                            <div className="listing-form-group">
                                <label htmlFor="duration">Auction Duration (starts now)</label>
                                <select
                                    id="duration"
                                    value={durationMinutes}
                                    onChange={(event) => setDurationMinutes(Number(event.target.value))}
                                >
                                    {AUCTION_DURATIONS.map((d) => (
                                        <option key={d.minutes} value={d.minutes}>{d.label}</option>
                                    ))}
                                </select>
                            </div>
                            <div className="listing-form-group">
                                <label htmlFor="reservePrice">Reserve Price (THB, Optional)</label>
                                <input
                                    id="reservePrice"
                                    type="number"
                                    min="1"
                                    step="0.01"
                                    value={reservePrice}
                                    onChange={(event) => setReservePrice(event.target.value)}
                                    placeholder="Optional minimum acceptable price"
                                />
                            </div>
                        </>
                    )}
                </section>
                {message && (
                    <p className="listing-form-message" role="status">
                        {message}
                    </p>
                )}
                <div className="create-listing-actions">
                    <button type="button" className="save-draft-button" onClick={handleSaveDraft} disabled={busy}>
                        Save Draft
                    </button>
                    <button type="submit" className="publish-listing-button" disabled={busy}>
                        {busy ? "Publishing..." : "Publish Listing"}
                    </button>
                </div>
            </form>
        </main>
    );
}
