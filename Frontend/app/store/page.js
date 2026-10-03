import Link from "next/link";
import { getListings, PAGE_SIZE_OPTIONS } from "../listings";

function parsePositiveInteger(value, fallback) {
  const parsed = Number(Array.isArray(value) ? value[0] : value);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : fallback;
}

function getSingleParam(value) {
  return Array.isArray(value) ? value[0] : value;
}

function parseOptionalPrice(value, label) {
  const rawValue = getSingleParam(value);
  if (rawValue === undefined || rawValue.trim() === "") {
    return { value: undefined, error: null };
  }

  const price = Number(rawValue);
  if (!Number.isFinite(price) || price < 0) {
    return { value: undefined, error: `${label} must be a non-negative number.` };
  }

  return { value: price, error: null };
}

export default async function StorePage({ searchParams }) {
  const params = await searchParams;
  const requestedPageSize = parsePositiveInteger(params.pageSize, 20);
  const pageSize = PAGE_SIZE_OPTIONS.includes(requestedPageSize)
    ? requestedPageSize
    : 20;
  const page = parsePositiveInteger(params.page, 1);
  const minPriceResult = parseOptionalPrice(params.minPrice, "Minimum price");
  const maxPriceResult = parseOptionalPrice(params.maxPrice, "Maximum price");
  const priceError = minPriceResult.error
    || maxPriceResult.error
    || (minPriceResult.value !== undefined
      && maxPriceResult.value !== undefined
      && minPriceResult.value > maxPriceResult.value
      ? "Minimum price cannot be greater than maximum price."
      : null);
  const minPrice = minPriceResult.value;
  const maxPrice = maxPriceResult.value;
  const result = priceError
    ? null
    : await getListings({ page, pageSize, minPrice, maxPrice });
  const filterQuery = {
    ...(minPrice !== undefined ? { minPrice } : {}),
    ...(maxPrice !== undefined ? { maxPrice } : {}),
  };

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
          {/* price filter */}
          <form className="price-filter" action="/store" method="get">
            <h2>Price range</h2>
            <input type="hidden" name="pageSize" value={pageSize} />
            <label htmlFor="min-price">Min price</label>
            <input
              id="min-price"
              name="minPrice"
              type="number"
              min="0"
              step="0.01"
              placeholder="No minimum"
              defaultValue={getSingleParam(params.minPrice) ?? ""}
            />
            <label htmlFor="max-price">Max price</label>
            <input
              id="max-price"
              name="maxPrice"
              type="number"
              min="0"
              step="0.01"
              placeholder="No maximum"
              defaultValue={getSingleParam(params.maxPrice) ?? ""}
            />
            <button type="submit">Apply price range</button>
          </form>
        </nav>

        {/* listing */}
        <section className="store-listing" aria-label="Gem listings">
          {/* setting for listing per page */}
          <form className="listing-size" action="/store" method="get">
            <label htmlFor="page-size">Listings per page </label>
            {minPrice !== undefined && (
              <input type="hidden" name="minPrice" value={minPrice} />
            )}
            {maxPrice !== undefined && (
              <input type="hidden" name="maxPrice" value={maxPrice} />
            )}
            <select id="page-size" name="pageSize" defaultValue={pageSize}>
              {PAGE_SIZE_OPTIONS.map((size) => (
                <option key={size} value={size}>{size}</option>
              ))}
            </select>
            <button type="submit">Apply</button>
          </form>
          {/* items */}
          {priceError ? (
            <p className="price-filter-error" role="alert">{priceError}</p>
          ) : (
            <>
              <div id="items">
                {result.items.length > 0 ? result.items.map((item) => (
                  <Link
                    className="store-item-link"
                    href={{ pathname: "/item", query: { id: item.id } }}
                    key={item.id}
                  >
                    <article className="store-item">
                      <img src={item.image} alt={item.title} />
                      <h2>{item.title}</h2>
                      <p>{item.description}</p>
                      <p className="store-item-price">
                        {new Intl.NumberFormat("en-US", {
                          style: "currency",
                          currency: "USD",
                        }).format(item.price)}
                      </p>
                    </article>
                  </Link>
                )) : (
                  <p className="empty-listings">No listings match your current filters.</p>
                )}
              </div>
              <nav className="pages" aria-label="Listing pages">
                {result.page > 1 && (
                  <Link href={{
                    pathname: "/store",
                    query: { page: result.page - 1, pageSize, ...filterQuery },
                  }}>
                    Previous
                  </Link>
                )}
                <span> Page {result.page} of {result.totalPages} ({result.total} listings) </span>
                {result.page < result.totalPages && (
                  <Link href={{
                    pathname: "/store",
                    query: { page: result.page + 1, pageSize, ...filterQuery },
                  }}>
                    Next
                  </Link>
                )}
              </nav>
            </>
          )}
        </section>
      </div>
    </main>
  );
}