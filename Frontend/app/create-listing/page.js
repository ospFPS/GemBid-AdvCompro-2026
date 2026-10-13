
"use client";

import { useState } from "react";
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

    function handleSaveDraft() {
        setMessage("Draft saved for demonstration. It has not been stored permanently.");
    }
    function handlePublish(event) {
        event.preventDefault();
        setMessage("Listing form completed for demonstration. Backend publishing is not connected yet.");
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
                    <h2>Gemstone Images </h2>
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
                            <p className="selected-file-info">
                                {images.length} image(s) selected
                            </p>
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
                    <button type="button" className="save-draft-button" onClick={handleSaveDraft}>
                        Save Draft
                    </button>
                    <button type="submit" className="publish-listing-button">
                        Publish Listing
                    </button>
                </div>
            </form>
        </main>
    );
}
