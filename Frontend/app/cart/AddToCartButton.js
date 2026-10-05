"use client";

import { useState } from "react";
import { addCartItem } from "./cartStorage";

export default function AddToCartButton({ itemId }) {
	const [message, setMessage] = useState("");

	function handleAddToCart() {
		try {
			addCartItem(itemId);
			setMessage("Added to cart.");
		} catch (error) {
			setMessage(error instanceof Error ? error.message : "Unable to update the cart.");
		}
	}

	return (
		<>
			<button type="button" onClick={handleAddToCart}>Add to cart</button>
			{message && <span role="status">{message}</span>}
		</>
	);
}
