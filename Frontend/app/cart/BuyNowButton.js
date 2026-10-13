"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { addCartItem } from "./cartStorage";

export default function BuyNowButton({ itemId }) {
	const router = useRouter();
	const [message, setMessage] = useState("");

	function handleBuyNow() {
		try {
			addCartItem(itemId);
			router.push("/cart");
		} catch (error) {
			setMessage(error instanceof Error ? error.message : "Unable to update the cart.");
		}
	}

	return (
		<>
			<button
				className="item-action-button item-action-button--primary"
				type="button"
				onClick={handleBuyNow}
			>
				Buy now
			</button>
			{message && <span role="alert">{message}</span>}
		</>
	);
}
