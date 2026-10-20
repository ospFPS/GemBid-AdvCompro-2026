"use client";

import { useEffect, useState } from "react";
import { hasUnreadNotifications as checkUnread } from "../lib/api";

export default function Home() {
    const [hasUnreadNotifications, setHasUnreadNotifications] = useState(false);

    useEffect(() => {
        checkUnread().then(setHasUnreadNotifications);
    }, []);

    return (
        <main className="user-home">
            <header className="home-header">
                <h1>GemBid</h1>
                <nav>
                    <a href="/store">Marketplace</a>
                    <a href="/my-bids">My Bids</a>
                    <a href="/my-purchases">My Purchases</a>

                    <a
                        href="/notification"
                        className={hasUnreadNotifications ? "nav-link--unread" : ""}
                        onClick={() => setHasUnreadNotifications(false)}
                    >
                        Notifications
                        {hasUnreadNotifications && (
                            <span className="notification-dot"
                                aria-label="Unread notifications">
                                !
                            </span>
                        )}
                    </a>

                    <a href="/profile">Profile</a>
                    <a href="/become-seller?from=home" className="Become-seller">
                        Become a Seller
                    </a>
                </nav>
            </header>

            <main className="home-content">
                <section className="welcome-section">
                    <h2>Welcome to GemBid</h2>
                    <p>
                        Discover gemstones, shop from sellers,
                        and participate in auctions.
                    </p>
                </section>

                <section className="home-options">
                    <a href="/store" className="home-card">
                        <h3>Marketplace</h3>
                        <p>Browse gemstones available for purchase.</p>
                    </a>

                    <a href="/my-bids" className="home-card">
                        <h3>My Bids</h3>
                        <p>View your current and previous bids.</p>
                    </a>

                    <a href="/my-purchases" className="home-card">
                        <h3>My Purchases</h3>
                        <p>View gemstones you have purchased.</p>
                    </a>

                    <a
                        href="/notification"
                        className="home-card"
                        onClick={() => setHasUnreadNotifications(false)}
                    >
                        <h3>
                            Notifications
                            {hasUnreadNotifications && (
                                <span className="notification-dot"
                                    aria-label="Unread notifications">
                                    !
                                </span>
                            )}
                        </h3>
                        <p>
                            Check updates about your account and purchases.
                        </p>
                    </a>

                    <a href="/profile" className="home-card">
                        <h3>Profile</h3>
                        <p>View and manage your account information.</p>
                    </a>

                    <a
                        href="/become-seller?from=home"
                        className="home-card seller-card"
                    >
                        <h3>Become a Seller</h3>
                        <p>Apply to become a verified gemstone seller.</p>
                    </a>
                </section>
            </main>
        </main>
    );
}
