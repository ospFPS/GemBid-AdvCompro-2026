// reference
import "./globals.css";

// mata data
export const metadata = {
  title: "GemBid",
  description: "Browse gems on GemBid.",
};

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}