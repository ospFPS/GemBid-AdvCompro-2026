"use client";

import { useSearchParams } from "next/navigation";
export default function BecomeSeller() {
    const searchParams = useSearchParams();
    const from = searchParams.get("from");
    const backLink = from === "profile" ? "/profile" : "/";
    const backText = from === "profile"
        ? "← Back to Profile"
        : "← Back to Home";
    return (
        <main className="seller-page">
            <a href={backLink} className="seller-home-link">
                {backText}
            </a>
            <h1>Become a Seller</h1>
            <p className="seller-introduction">
                Apply to become a verified gemstone seller on GemBid.
            </p>
            <form className="seller-form">
                <section className="seller-form-card">
                    <h2>Seller Information</h2>
                    <label htmlFor="shopName">
                        Seller / Shop Name
                    </label>
                    <input
                        type="text"
                        id="shopName"
                        name="shopName"
                        required
                    />
                    <label htmlFor="legalName">
                        Legal Name
                    </label>
                    <input
                        type="text"
                        id="legalName"
                        name="legalName"
                        required
                    />
                    <label htmlFor="phone">
                        Phone Number
                    </label>
                    <input
                        type="tel"
                        id="phone"
                        name="phone"
                        required
                    />
                    <label htmlFor="address">
                        Address
                    </label>
                    <textarea
                        id="address"
                        name="address"
                        rows="4"
                        required
                    />
                    <label htmlFor="sellerType">
                        Seller Type
                    </label>
                    <select
                        id="sellerType"
                        name="sellerType"
                        required
                    >
                        <option value="">
                            Select seller type
                        </option>
                        <option value="individual">
                            Individual
                        </option>
                        <option value="business">
                            Business / Shop
                        </option>
                    </select>
                    <label htmlFor="description">
                        Seller Description
                    </label>
                    <textarea
                        id="description"
                        name="description"
                        rows="5"
                    />
                </section>
                <section className="seller-form-card">
                    <h2>Documents</h2>
                    <label htmlFor="identification">
                        Identification Document
                    </label>
                    <input
                        type="file"
                        id="identification"
                        name="identification"
                    />
                    <label htmlFor="businessDocument">
                        Business / Shop Document
                    </label>
                    <input
                        type="file"
                        id="businessDocument"
                        name="businessDocument"
                    />
                </section>
                <section className="seller-form-card">
                    <h2>Application</h2>
                    <p>
                        Your application will be reviewed by an
                        administrator before you can sell gemstones.
                    </p>
                    <button type="submit">
                        Submit Seller Application
                    </button>
                </section>
            </form>
        </main>
    );
}