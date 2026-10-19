"use client";
import { useEffect, useState } from "react";
import { api, formatDate, logout, saveLogin, getToken } from "../lib/api";

const SELLER_STATUS_TEXT = {
    NONE: "You are currently a general user.",
    PENDING: "Your seller application is waiting for admin review.",
    VERIFIED: "You are a verified seller.",
    REJECTED: "Your seller application was rejected.",
    SUSPENDED: "Your seller account is suspended.",
};

export default function Profile() {
    const [editing, setEditing] = useState(false);
    const [user, setUser] = useState(null);
    const [plan, setPlan] = useState(null);
    const [message, setMessage] = useState("");
    const [planError, setPlanError] = useState("");
    const [loading, setLoading] = useState(true);
    const [busy, setBusy] = useState(false);

    useEffect(() => {
        if (!getToken()) {
            window.location.replace("/login");
            return;
        }
        loadProfile();
    }, []);

    function loadProfile() {
        setLoading(true);
        setMessage("");
        setPlanError("");
        api("/api/me")
            .then(setUser)
            .catch((err) => setMessage(`Unable to load your profile: ${err.message}`))
            .finally(() => setLoading(false));
        api("/api/subscriptions/me")
            .then(setPlan)
            .catch((err) => setPlanError(`Subscription details are unavailable: ${err.message}`));
    }

    async function saveChanges(event) {
        event.preventDefault();
        const form = new FormData(event.currentTarget);
        setBusy(true);
        setMessage("");
        try {
            const result = await api("/api/me", {
                method: "PUT",
                body: {
                    username: form.get("username"),
                    full_name: form.get("full_name"),
                    phone: form.get("phone"),
                },
            });
            if (form.get("new_password")) {
                await api("/api/me/password", {
                    method: "PUT",
                    body: {
                        current_password: form.get("current_password"),
                        new_password: form.get("new_password"),
                    },
                });
            }
            saveLogin(getToken(), result.user);
            setUser(result.user);
            setEditing(false);
            setMessage("Profile saved.");
        } catch (err) {
            setMessage(err.message);
        }
        setBusy(false);
    }

    async function upgrade() {
        setBusy(true);
        try {
            await api("/api/subscriptions/upgrade", { method: "POST", body: { payment_method: "MOCK_CARD" } });
            setPlan(await api("/api/subscriptions/me"));
            setMessage("You are now on GemBid PRO.");
        } catch (err) {
            setMessage(err.message);
        }
        setBusy(false);
    }

    if (!user) {
        return (
            <main className="profile-page">
                {loading ? (
                    <p role="status">Loading profile...</p>
                ) : (
                    <>
                        <p role="alert">{message}</p>
                        <button type="button" onClick={loadProfile}>Retry</button>
                        {" "}
                        <a href="/store" className="profile-home-link">Back to Store</a>
                    </>
                )}
            </main>
        );
    }

    return (
        <main className="profile-page">
            <a href="/store" className="profile-home-link">
                ← Back to Store
            </a>
            <h1>Profile</h1>
            {message && <p role="status">{message}</p>}
            <section className="profile-card">
                <h2>Account Information</h2>
                {editing ? (
                    <form className="profile-edit-form" onSubmit={saveChanges}>
                        <label htmlFor="username">Username</label>
                        <input type="text" id="username" name="username" defaultValue={user.username} />
                        <label htmlFor="name">Full Name</label>
                        <input type="text" id="name" name="full_name" defaultValue={user.full_name} required />
                        <label htmlFor="email">Email</label>
                        <input type="email" id="email" defaultValue={user.email ?? ""} disabled />
                        <label htmlFor="phone">Phone Number</label>
                        <input type="tel" id="phone" name="phone" defaultValue={user.phone ?? ""} />
                        <label htmlFor="current_password">Current Password (only to change password)</label>
                        <input type="password" id="current_password" name="current_password" />
                        <label htmlFor="password">New Password</label>
                        <input
                            type="password"
                            id="password"
                            name="new_password"
                            placeholder="Leave empty to keep"
                            minLength={8}
                            maxLength={72}
                        />
                        <div className="profile-edit-actions">
                            <button type="submit" disabled={busy}>
                                {busy ? "Saving..." : "Save Changes"}
                            </button>
                            <button type="button" onClick={() => setEditing(false)}>
                                Cancel
                            </button>
                        </div>
                    </form>
                ) : (
                    <div className="profile-information">
                        <p><strong>Username:</strong> {user.username}</p>
                        <p><strong>Full Name:</strong> {user.full_name}</p>
                        <p><strong>Email:</strong> {user.email ?? "-"}</p>
                        <p><strong>Phone Number:</strong> {user.phone ?? "-"}</p>
                    </div>
                )}
                <p>
                    <strong>Account Type:</strong>{" "}
                    {user.role === "ADMIN" ? "Admin" : user.seller_status === "VERIFIED" ? "Seller" : "Normal User"}
                </p>
            </section>
            <section className="profile-card">
                <h2>User Status</h2>
                <div className="profile-status">
                    <p>{SELLER_STATUS_TEXT[user.seller_status] ?? user.seller_status}</p>
                </div>
                {user.seller_status === "NONE" && (
                    <a href="/become-seller?from=profile" className="profile-link">
                        Become a Seller
                    </a>
                )}
            </section>
            <section className="profile-card">
                <h2>Plan</h2>
                <div className="profile-status">
                    {planError ? (
                        <p role="alert">{planError}</p>
                    ) : plan?.plan === "PRO" ? (
                        <p>GemBid PRO until {formatDate(plan.subscription.end_date)}: unlimited live listings.</p>
                    ) : (
                        <p>FREE plan: up to 5 live listings as a seller.</p>
                    )}
                </div>
                <button type="button" onClick={upgrade} disabled={busy}>
                    {plan?.plan === "PRO" ? "Extend PRO (+30 days, ฿299)" : "Upgrade to PRO (฿299 / 30 days)"}
                </button>
            </section>
            <section className="profile-card">
                <h2>Account Settings</h2>
                <div className="profile-actions">
                    <button onClick={() => setEditing(true)}>
                        Edit Profile
                    </button>
                    <button
                        onClick={() => {
                            logout();
                            window.location.href = "/";
                        }}
                    >
                        Log Out
                    </button>
                </div>
            </section>
        </main>
    );
}
