"use client";

import { useState } from "react";
import { api } from "../lib/api";

// Creates a real order for this listing, then opens the order page to pay.
export default function BuyNowButton({ itemId }) {
	const [message, setMessage] = useState("");
	const [busy, setBusy] = useState(false);

	async function handleBuyNow() {
		setBusy(true);
		try {
			const order = await api(`/api/listings/${itemId}/buy`, { method: "POST" });
			window.location.href = `/my-purchases/view-order/${order.order_id}`;
		} catch (error) {
			setMessage(error.message);
			setBusy(false);
		}
	}

	return (
		<>
			<button
				className="item-action-button item-action-button--primary"
				type="button"
				onClick={handleBuyNow}
				disabled={busy}
			>
				{busy ? "Placing order..." : "Buy now"}
			</button>
			{message && <span role="alert">{message}</span>}
		</>
	);
}
