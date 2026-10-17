import Link from "next/link";
import { notFound } from "next/navigation";
import { getListingById, getMoreListings } from "../listings";
import AddToCartButton from "../cart/AddToCartButton";
import BuyNowButton from "../cart/BuyNowButton";
import { formatTHB } from "../lib/api";

export default async function ItemPage({ searchParams }) {
	const { id } = await searchParams;
	const item = await getListingById(id);

	if (!item) {
		notFound();
	}

	const moreItems = await getMoreListings(item.id, 5);

	return (
		<main className="item-detail-layout">
			<Link className="item-back-link" href="/store">
				<span aria-hidden="true">←</span>
				Back to marketplace
			</Link>
			<article className="item-detail">
				<div className="item-detail-imageBox">
					<img src={item.image} alt={item.title} />
				</div>
				<div className="item-detail-infoBox">
					<div>
						{/* check item selling method (buy / bid) */}
						<span className={`listing-type-badge ${item.isBid ? "listing-type-badge--bid" : "listing-type-badge--fixed"}`}>
							{item.isBid ? "Bidding required" : "Buy it now"}
						</span>
						<h1>{item.title}</h1>
						<p>{item.description}</p>
						{/* seller detail */}
						<div className="item-detail-seller">
							<img src={item.pfp} alt="" />
							<div>
								<span className="seller-label">Listed by</span>
								<strong>{item.seller}</strong>
							</div>
						</div>
						<div className="item-detail-price-box">
							<span>Price</span>
							<p className="item-detail-price">
								{formatTHB(item.price)}
							</p>
						</div>
						{/* button for bid or buy */}
						<div className="item-detail-actions">
							{item.isBid ? (
								<Link
									className="item-action-button item-action-button--primary"
									href={{ pathname: "/bid", query: { itemId: item.id } }}
								>
									Bid now
								</Link>
							) : (
								<>
									<BuyNowButton itemId={item.id} />
									<AddToCartButton itemId={item.id} />
									<Link className="item-action-button item-action-button--secondary" href="/cart">View cart</Link>
								</>
							)}
						</div>
					</div>
				</div>
			</article>
			{/* more items */}
			<section className="more-items" aria-labelledby="more-items-heading">
				<h2 id="more-items-heading">More items</h2>
				<div className="more-items-grid">
					{moreItems.map((moreItem) => (
						<Link
							className="store-item-link"
							href={{ pathname: "/item", query: { id: moreItem.id } }}
							key={moreItem.id}
						>
							<article className="store-item">
								<img src={moreItem.image} alt={moreItem.title} />
								<span className={`listing-type-badge ${moreItem.isBid ? "listing-type-badge--bid" : "listing-type-badge--fixed"}`}>
									{moreItem.isBid ? "Bidding required" : "Buy it now"}
								</span>
								<h2>{moreItem.title}</h2>
								<p className="store-item-description">{moreItem.description}</p>
								{/* seller detail */}
								<div className="listing-seller">
									<img src={moreItem.pfp} alt="" />
									<span>Sold by {moreItem.seller}</span>
								</div>
								<p className="store-item-price">
									{formatTHB(moreItem.price)}
								</p>
							</article>
						</Link>
					))}
				</div>
			</section>
		</main>
	);
}
