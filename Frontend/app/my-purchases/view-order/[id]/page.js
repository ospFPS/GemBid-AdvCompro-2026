import { purchases } from "../../../data/purchases";
function getProgressSteps(status) {
    const progressSteps = [
        "Order Created",
        "Payment",
        "Seller Preparing Item",
        "Shipped",
        "Delivered",
        "Completed",
    ];
    const statusIndex = progressSteps.indexOf(status);
    return progressSteps.map((step, index) => ({
        name: step,
        completed: index <= statusIndex,
    }));
}
export default async function ViewOrder({ params, searchParams }) {
    const { id } = await params;
    const { from } = await searchParams;
    const purchase = purchases.find((purchase) => purchase.id === Number(id));
    const progressSteps = [
        {
            name: "Order Created",
            completed: true,
        },
        {
            name: "Payment",
            completed: true,
        },
        {
            name: "Seller Preparing Item",
            completed: purchase.status !== "Processing",
        },
        {
            name: "Shipped",
            completed: purchase.status === "Shipped" ||
                    purchase.status === "Delivered" ||
                    purchase.status === "Completed",
        },
        {
            name: "Delivered",
            completed: purchase.status === "Delivered" ||
                    purchase.status === "Completed",
        },
        {
            name: "Completed",
            completed: purchase.status === "Completed",
        }
    ];
    return (
        <main className="view-order-page">
            <div className="view-order-sticky-nav">
                <a href="/" className="view-order-home">
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
            <h1>Order Details</h1>
            <section className="order-details">
                <h2>{purchase.gemstone}</h2>
                <p><strong>Order ID:</strong> {purchase.orderId}</p>
                <img src={purchase.image} alt={purchase.gemstone}/>
                <p><strong>Purchase type:</strong> {purchase.purchaseType}</p>
                <p><strong>Price:</strong> {purchase.price}</p>
                <p><strong>Purchase date:</strong> {purchase.purchaseDate}</p>
                <p><strong>Status:</strong> {purchase.status}</p>
            </section>
            <section className="order-progress">
                <h2>Order Progress</h2>
                {progressSteps.map((step) => (
                    <p key={step.name} className={step.completed ? "completed" : "not-completed"}>
                        {step.completed ? "✓" : "○"} {step.name}
                    </p>
                ))}
            </section>
        </main>
    );
}