"use client";

import Link from "next/link";
import { useState, useSyncExternalStore } from "react";
import { api, formatTHB } from "../lib/api";
import {
	getCartSnapshot,
	getServerCartSnapshot,
	parseCartSnapshot,
	readCartIds,
	removeCartItem,
	saveCartIds,
	subscribeToCart,
} from "./cartStorage";

export default function CartContents({ listings }) {
	const cartSnapshot = useSyncExternalStore(
		subscribeToCart,
		getCartSnapshot,
		getServerCartSnapshot
	);
	const { ids: cartIds, error: storageError } = parseCartSnapshot(cartSnapshot);
	const [actionError, setActionError] = useState("");
	const [checkingOut, setCheckingOut] = useState(false);

	function updateCart(action, itemId) {
		try {
			if (action === "add") {
				saveCartIds([...cartIds, itemId]);
			} else {
				removeCartItem(itemId);
			}
			setActionError("");
		} catch (updateError) {
			setActionError(updateError instanceof Error ? updateError.message : "Unable to update the cart.");
		}
	}

	// Each listing is one unique gem, so checkout buys every listing in the cart once.
	async function checkout() {
		setCheckingOut(true);
		const failed = [];
		for (const { item } of cartItems) {
			try {
				await api(`/api/listings/${item.id}/buy`, { method: "POST" });
				saveCartIds(readCartIds().filter((id) => id !== item.id));
			} catch (err) {
				failed.push(`${item.title}: ${err.message}`);
			}
		}
		if (failed.length === 0) {
			window.location.href = "/my-purchases";
			return;
		}
		setActionError(`Some items could not be bought. ${failed.join(" ")}`);
		setCheckingOut(false);
	}

	function clearInvalidCart() {
		try {
			saveCartIds([]);
			setActionError("");
		} catch (updateError) {
			setActionError(updateError instanceof Error ? updateError.message : "Unable to clear the cart.");
		}
	}

	const quantities = new Map();
	for (const itemId of cartIds) {
		quantities.set(itemId, (quantities.get(itemId) ?? 0) + 1);
	}

	const cartItems = Array.from(quantities, ([id, quantity]) => ({
		item: listings.find((listing) => listing.id === id),
		quantity,
	})).filter(({ item }) => item);
	const itemCount = cartItems.reduce((count, entry) => count + entry.quantity, 0);
	const totalPrice = cartItems.reduce(
		(total, { item, quantity }) => total + item.price * quantity,
		0
	);

	return (
		<>
			{(storageError || actionError) && (
				<div className="cart-error" role="alert">
					<p>{actionError || storageError}</p>
					{storageError && storageError !== "Cart storage is unavailable in this browser." && (
						<button type="button" onClick={clearInvalidCart}>Clear invalid cart data</button>
					)}
				</div>
			)}
			<div className="cart-layout">
				<section className="cart-items" aria-label="Cart items">
					{cartItems.length > 0 ? cartItems.map(({ item, quantity }) => (
						<article className="cart-item" key={item.id}>
							<Link
								className="cart-item-link"
								href={{ pathname: "/item", query: { id: item.id } }}
							>
								<img src={item.image} alt={item.title} />
								<span className="cart-item-details">
									<strong>{item.title}</strong>
									<span>{formatTHB(item.price)}</span>
								</span>
							</Link>
							<div className="cart-item-controls" aria-label={`${item.title} quantity`}>
								<button
									type="button"
									onClick={() => updateCart("remove", item.id)}
									aria-label={`Remove one ${item.title} from cart`}
								>
									-
								</button>
								<span>{quantity}</span>
								<button
									type="button"
									onClick={() => updateCart("add", item.id)}
									aria-label={`Add one ${item.title} to cart`}
								>
									+
								</button>
							</div>
						</article>
					)) : (
						<p className="cart-empty">Your cart is empty.</p>
					)}
				</section>
				<aside className="cart-summary" aria-labelledby="cart-summary-heading">
					<h2 id="cart-summary-heading">Order summary</h2>
					<p>
						<span>Items ({itemCount})</span>
						<strong>{formatTHB(totalPrice)}</strong>
					</p>
					<p className="cart-summary-total">
						<span>Total</span>
						<strong>{formatTHB(totalPrice)}</strong>
					</p>
					<button
						className="item-action-button"
						type="button"
						onClick={checkout}
						disabled={checkingOut || cartItems.length === 0}
					>
						{checkingOut ? "Placing orders..." : "Checkout"}
					</button>
				</aside>
			</div>
			<Link className="nav-link cart-continue" href="/store">Continue shopping</Link>
		</>
	);
}
