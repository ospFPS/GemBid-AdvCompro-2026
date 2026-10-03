import Link from "next/link";
import { notFound } from "next/navigation";
import { getListingById } from "../listings";

export default async function ItemPage({ searchParams }) {
	const { id } = await searchParams;
	const item = await getListingById(id);

	if (!item) {
		notFound();
	}

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
						<p className="item-detail-price">
							{new Intl.NumberFormat("en-US", {
								style: "currency",
								currency: "USD",
							}).format(item.price)}
						</p>
						<div className="item-detail-actions">
							<button type="button">Buy now</button>
							<button type="button">Add to cart</button>
						</div>
					</div>
				</div>
			</article>
		</main>
	);
}
