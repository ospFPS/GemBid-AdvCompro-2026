import Link from "next/link";
import { notFound } from "next/navigation";
import { getListingById } from "../listings";
import { formatTHB } from "../lib/api";
import BidForm from "./BidForm";

export default async function BidPage({ searchParams }) {
	const { itemId } = await searchParams;
	const item = await getListingById(itemId);

	if (!item || !item.isBid || !item.auctionId) {
		notFound();
	}

	return (
		<main className="cart-page">
			<Link className="nav-link" href={{ pathname: "/item", query: { id: item.id } }}>
				Back to listing
			</Link>
			<h1>Place a bid</h1>
			<div className="bid-layout">
				<section className="bid-listing" aria-labelledby="bid-listing-heading">
					<img className="bid-listing-image" src={item.image} alt={item.title} />
					<div className="bid-listing-details">
						<span className="listing-type-badge listing-type-badge--bid">
							Bidding required
						</span>
						<h2 id="bid-listing-heading">{item.title}</h2>
						<p>{item.description}</p>
						<p className="bid-listing-price">
							<span>Current price</span>
							<strong>{formatTHB(item.price)}</strong>
						</p>
						<div className="listing-seller">
							<img src={item.pfp} alt="" />
							<span>Sold by {item.seller}</span>
						</div>
					</div>
				</section>

				<aside className="bid-summary" aria-labelledby="bid-summary-heading">
					<h2 id="bid-summary-heading">Your bid</h2>
					<BidForm auctionId={item.auctionId} />
				</aside>
			</div>
			<Link className="nav-link cart-continue" href="/store">
				Continue shopping
			</Link>
		</main>
	);
}