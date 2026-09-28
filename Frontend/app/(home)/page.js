export default function Home() {
    return (
        <main className="user-home">
            <header className="home-header">
                <h1>GemBid</h1>
                <nav>
                    <a href="/store">Marketplace</a>
                    <a href="/auctions">Auctions</a>
                    <a href="/my-bids">My Bids</a>
                    <a href="/my-purchases">My Purchases</a>
                    <a href="/notifications">Notifications</a>
                    <a href="/profile">Profile</a>
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
                        <p>
                            Browse gemstones available for purchase.
                        </p>
                    </a>

                    <a href="/auctions" className="home-card">
                        <h3>Auctions</h3>
                        <p>
                            View ongoing and upcoming gemstone auctions.
                        </p>
                    </a>

                    <a href="/my-bids" className="home-card">
                        <h3>My Bids</h3>
                        <p>
                            View your current and previous bids.
                        </p>
                    </a>

                    <a href="/my-purchases" className="home-card">
                        <h3>My Purchases</h3>
                        <p>
                            View gemstones you have purchased.
                        </p>
                    </a>

                    <a href="/notifications" className="home-card">
                        <h3>Notifications</h3>
                        <p>
                            Check updates about your account and purchases.
                        </p>
                    </a>

                    <a href="/profile" className="home-card">
                        <h3>Profile</h3>
                        <p>
                            View and manage your account information.
                        </p>
                    </a>

                    <a
                        href="/become-seller"
                        className="home-card seller-card"
                    >
                        <h3>Become a Seller</h3>
                        <p>
                            Apply to become a verified gemstone seller.
                        </p>
                    </a>
                </section>
            </main>
        </main>
    );
}