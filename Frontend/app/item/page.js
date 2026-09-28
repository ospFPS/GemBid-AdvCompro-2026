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
				<img src={item.image} alt={item.title} />
				<div>
					<h1>{item.title}</h1>
					<p>{item.description}</p>
				</div>
			</article>
		</main>
	);
}
