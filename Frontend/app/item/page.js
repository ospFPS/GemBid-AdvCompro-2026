import Link from "next/link";
import { notFound } from "next/navigation";
import { getListingById, getMoreListings } from "../listings";
import AddToCartButton from "../cart/AddToCartButton";

export default async function ItemPage({ searchParams }) {
	const { id } = await searchParams;
	const item = await getListingById(id);

	if (!item) {
		notFound();
	}

	const moreItems = await getMoreListings(item.id, 5);

	return (
		<main className="item-detail-layout">
			<Link className="nav-link" href="/store">
				Back to marketplace
			</Link>
			<article className="item-detail">
				<div className="item-detail-imageBox">
					<div>
						<img src={item.image} alt={item.title} />
					</div>
				</div>
				<div className="item-detail-infoBox">
					<div>
						<h1>{item.title}</h1>
						<p>{item.description}</p>
						{/* seller detail */}
						<div className="item-detail-seller">
							<img src={item.pfp} alt="" />
							<div>
								<span className="seller-label">Seller</span>
								<strong>{item.seller}</strong>
							</div>
						</div>
						<p className="item-detail-price">
							{new Intl.NumberFormat("en-US", {
								style: "currency",
								currency: "USD",
							}).format(item.price)}
						</p>
						<div className="item-detail-actions">
							<button type="button">Buy now</button>
							{/* cart action */}
							<AddToCartButton itemId={item.id} />
							<Link className="nav-link" href="/cart">View cart</Link>
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
								<h2>{moreItem.title}</h2>
								<p className="store-item-description">{moreItem.description}</p>
								{/* seller detail */}
								<div className="listing-seller">
									<img src={moreItem.pfp} alt="" />
									<span>Sold by {moreItem.seller}</span>
								</div>
								<p className="store-item-price">
									{new Intl.NumberFormat("en-US", {
										style: "currency",
										currency: "USD",
									}).format(moreItem.price)}
								</p>
							</article>
						</Link>
					))}
				</div>
			</section>
		</main>
	);
}
