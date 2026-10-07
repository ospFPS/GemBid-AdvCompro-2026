import Link from "next/link";
import CartContents from "./CartContents";
import { getAllListings } from "../listings";

export default async function CartPage() {
	const listings = await getAllListings();

	return (
		<main className="cart-page">
			<Link className="nav-link" href="/store">Back to marketplace</Link>
			<h1>Shopping cart</h1>
			{/* cart items and order total */}
			<CartContents listings={listings} />
		</main>
	);
}