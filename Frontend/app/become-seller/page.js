"use client";

import { Suspense, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { api } from "../lib/api";

const STATUS_TEXT = {
    PENDING: "Your application has been submitted and is waiting for admin review.",
    VERIFIED: "You are a verified seller. You can now create listings.",
    REJECTED: "Your seller application was rejected.",
    SUSPENDED: "Your seller account is suspended.",
};

function BecomeSellerForm() {
    const searchParams = useSearchParams();
    const [status, setStatus] = useState(null);
    const [error, setError] = useState("");
    const [busy, setBusy] = useState(false);

    useEffect(() => {
        api("/api/seller/status").then((data) => setStatus(data.seller_status)).catch((err) => setError(err.message));
    }, []);

    async function submitApplication(event) {
        event.preventDefault();
        const form = new FormData(event.currentTarget);
        setBusy(true);
        setError("");
        try {
            await api("/api/seller/apply", {
                method: "POST",
                body: {
                    shop_name: form.get("shopName"),
                    legal_name: form.get("legalName"),
                    seller_type: form.get("sellerType"),
                    address: form.get("address"),
                    description: form.get("description"),
                },
            });
            setStatus("PENDING");
        } catch (err) {
            setError(err.message);
        }
        setBusy(false);
    }

    const from = searchParams.get("from");
    const backLink = from === "profile" ? "/profile" : "/store";
    const backText = from === "profile"
        ? "← Back to Profile"
        : "← Back to Store";
    return (
        <main className="seller-page">
            <a href={backLink} className="seller-home-link">
                {backText}
            </a>
            <h1>Become a Seller</h1>
            <p className="seller-introduction">
                Apply to become a verified gemstone seller on GemBid.
            </p>
            {error && <p role="alert">{error}</p>}
            {status && status !== "NONE" ? (
            <section className="seller-form-card">
                <h2>Application Status: {status}</h2>
                <p>{STATUS_TEXT[status]}</p>
                {status === "VERIFIED" && <a href="/create-listing">Create your first listing</a>}
            </section>
            ) : (
            <form className="seller-form" onSubmit={submitApplication}>
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
                        <option value="Individual">
                            Individual
                        </option>
                        <option value="Business">
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
                    <button type="submit" disabled={busy}>
                        {busy ? "Submitting..." : "Submit Seller Application"}
                    </button>
                </section>
            </form>
            )}
        </main>
    );
}

// useSearchParams() needs a Suspense boundary for the production build.
export default function BecomeSeller() {
    return (
        <Suspense>
            <BecomeSellerForm />
        </Suspense>
    );
}
