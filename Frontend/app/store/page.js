import Link from "next/link";

const items = [
  {
    title: "Leek (This is a test item)",
    description: "A rare gem from the northern caves.",
    image: "/assets/Leek.jpg",
  },
  {
    title: "Amog (This is a test item)",
    description: "A gem from mars.",
    image: "/assets/Reaction.jpg",
  },
  {
    title: "Leek (This is a test item)",
    description: "A rare gem from the northern caves.",
    image: "/assets/Leek.jpg",
  },
  {
    title: "Amog (This is a test item)",
    description: "A gem from mars.",
    image: "/assets/Reaction.jpg",
  },
  {
    title: "Leek (This is a test item)",
    description: "A rare gem from the northern caves.",
    image: "/assets/Leek.jpg",
  },
  {
    title: "Amog (This is a test item)",
    description: "A gem from mars.",
    image: "/assets/Reaction.jpg",
  },
];

export default function StorePage() {
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
          <div id="items">
            {items.map((item, index) => (
              <article className="store-item" key={`${item.title}-${index}`}>
                <img src={item.image} alt={item.title} />
                <h2>{item.title}</h2>
                <p>{item.description}</p>
              </article>
            ))}
          </div>
        </section>
      </div>
    </main>
  );
}