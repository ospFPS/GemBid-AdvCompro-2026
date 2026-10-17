"use client";

import { useCallback, useEffect, useState } from "react";
import { api, formatDate, toPurchase } from "../../../lib/api";

const STEPS = ["Order Created", "Payment", "Seller Preparing Item", "Shipped", "Delivered"];
// How many steps are done for each backend order status.
const DONE = { PENDING_PAYMENT: 1, PAID: 2, SHIPPED: 4, DELIVERED: 5, CANCELLED: 1 };

export default function OrderView({ id }) {
	const [detail, setDetail] = useState(null);
	const [purchase, setPurchase] = useState(null);
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);

	const load = useCallback(async () => {
		try {
			const data = await api(`/api/orders/${id}`);
			setDetail(data);
			setPurchase(await toPurchase(data.order));
		} catch (err) {
			setError(err.message);
		}
	}, [id]);

	useEffect(() => {
		load();
	}, [load]);

	async function act(path, body) {
		setBusy(true);
		setError("");
		try {
			await api(path, { method: "POST", body });
			await load();
		} catch (err) {
			setError(err.message);
		}
		setBusy(false);
	}

	function pay(event) {
		event.preventDefault();
		const form = new FormData(event.currentTarget);
		act(`/api/orders/${id}/pay`, {
			payment_method: form.get("payment_method"),
			shipping_address: form.get("shipping_address"),
		});
	}

	if (!detail || !purchase) {
		return <p>{error || "Loading order..."}</p>;
	}

	const { order, shipping } = detail;
	const done = DONE[order.status] ?? 0;

	return (
		<>
			<h1>Order Details</h1>
			<section className="order-details">
				<h2>{purchase.gemstone}</h2>
				<p><strong>Order ID:</strong> {purchase.orderId}</p>
				<img src={purchase.image} alt={purchase.gemstone} />
				<p><strong>Purchase type:</strong> {purchase.purchaseType}</p>
				<p><strong>Price:</strong> {purchase.price}</p>
				<p><strong>Purchase date:</strong> {purchase.purchaseDate}</p>
				<p><strong>Status:</strong> {purchase.status}</p>
				{shipping && (
					<>
						<p><strong>Ship to:</strong> {shipping.shipping_address}</p>
						{shipping.tracking_number && (
							<p>
								<strong>Tracking:</strong> {shipping.courier} {shipping.tracking_number}
								{shipping.shipped_at && ` (shipped ${formatDate(shipping.shipped_at)})`}
							</p>
						)}
					</>
				)}
			</section>

			{error && <p role="alert">{error}</p>}

			{order.status === "PENDING_PAYMENT" && (
				<section className="order-details">
					<h2>Payment</h2>
					<form onSubmit={pay}>
						<label htmlFor="payment_method">Payment method</label>
						<select id="payment_method" name="payment_method" defaultValue="MOCK_CARD">
							<option value="MOCK_CARD">Card (demo)</option>
							<option value="MOCK_QR">PromptPay QR (demo)</option>
						</select>
						<label htmlFor="shipping_address">Shipping address</label>
						<textarea id="shipping_address" name="shipping_address" rows={3} required />
						<button type="submit" disabled={busy}>{busy ? "Paying..." : "Pay now"}</button>
					</form>
					<button type="button" disabled={busy} onClick={() => act(`/api/orders/${id}/cancel`)}>
						Cancel order
					</button>
				</section>
			)}

			{order.status === "SHIPPED" && (
				<button type="button" disabled={busy} onClick={() => act(`/api/orders/${id}/confirm-delivery`)}>
					I received my gem
				</button>
			)}

			<section className="order-progress">
				<h2>Order Progress</h2>
				{order.status === "CANCELLED" ? (
					<p className="not-completed">This order was cancelled.</p>
				) : (
					STEPS.map((step, index) => (
						<p key={step} className={index < done ? "completed" : "not-completed"}>
							{index < done ? "✓" : "○"} {step}
						</p>
					))
				)}
			</section>
		</>
	);
}
