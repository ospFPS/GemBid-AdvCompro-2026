import OrderView from "./OrderView";

export default async function ViewOrder({ params, searchParams }) {
    const { id } = await params;
    const { from } = await searchParams;
    return (
        <main className="view-order-page">
            <div className="view-order-sticky-nav">
                <a href="/store" className="view-order-home">
                    <button type="button">Home</button>
                </a>
                <div>
                    <a
                        href={from === "notification" ? "/notification" : "/my-purchases"}
                        className="view-order-back-link"
                    >
                        ← Back to {from === "notification" ? "Notifications" : "My Purchases"}
                    </a>
                </div>
            </div>
            <OrderView id={id} />
        </main>
    );
}