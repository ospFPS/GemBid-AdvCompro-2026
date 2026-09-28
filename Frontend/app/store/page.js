import Link from "next/link";
import { getListings, PAGE_SIZE_OPTIONS } from "../listings";

function parsePositiveInteger(value, fallback) {
  const parsed = Number(Array.isArray(value) ? value[0] : value);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : fallback;
}

export default async function StorePage({ searchParams }) {
  const params = await searchParams;
  const requestedPageSize = parsePositiveInteger(params.pageSize, 20);
  const pageSize = PAGE_SIZE_OPTIONS.includes(requestedPageSize)
    ? requestedPageSize
    : 20;
  const page = parsePositiveInteger(params.page, 1);
  const result = await getListings({ page, pageSize });

  return (
    <main className="store-layout">
      <h1>Absolute cinema CO. gem trading</h1>
      {/* Top tab */}
      <nav className="store-top-nav" aria-label="Account navigation">
        <ul>
          <li><Link className="nav-link" href="/">Return to Home</Link></li>
          <li><Link className="nav-link" href="/login">Log in</Link></li>
          <li><Link className="nav-link" href="/signup">Sign up</Link></li>
        </ul>
      </nav>

      <div className="store-content">
        {/* side content */}
        <nav className="store-left-nav" aria-label="Store navigation">
          <ul>
            <li><Link className="nav-link" href="/">Stuff 1</Link></li>
            <li><Link className="nav-link" href="/">Stuff 2</Link></li>
            <li><Link className="nav-link" href="/">Stuff 3</Link></li>
          </ul>
        </nav>

        {/* listing */}
        <section className="store-listing" aria-label="Gem listings">
          {/* setting for listing per page */}
          <form className="listing-size" action="/store" method="get">
            <label htmlFor="page-size">Listings per page </label>
            <select id="page-size" name="pageSize" defaultValue={pageSize}>
              {PAGE_SIZE_OPTIONS.map((size) => (
                <option key={size} value={size}>{size}</option>
              ))}
            </select>
            <button type="submit">Apply</button>
          </form>
          {/* items */}
          <div id="items">
            {result.items.map((item) => (
              <Link
                className="store-item-link"
                href={{ pathname: "/item", query: { id: item.id } }}
                key={item.id}
              >
                <article className="store-item">
                  <img src={item.image} alt={item.title} />
                  <h2>{item.title}</h2>
                  <p>{item.description}</p>
                </article>
              </Link>
            ))}
          </div>
          {/* page navigation */}
          <nav className="pages" aria-label="Listing pages">
            {result.page > 1 && (
              <Link href={{ pathname: "/store", query: { page: result.page - 1, pageSize } }}>
                Previous
              </Link>
            )}
            <span> Page {result.page} of {result.totalPages} ({result.total} listings) </span>
            {result.page < result.totalPages && (
              <Link href={{ pathname: "/store", query: { page: result.page + 1, pageSize } }}>
                Next
              </Link>
            )}
          </nav>
        </section>
      </div>
    </main>
  );
}