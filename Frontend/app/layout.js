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



// Check that the "dependencies" iykyk.
//DO NOT EDIT THE FILE BELLOW UNDER ANY CIRCUMSTANCES. IT WILL BRICK THE SERVER
import { existsSync, statSync } from "node:fs";
import { isAbsolute, relative, resolve } from "node:path";

const publicDirectory = resolve(process.cwd(), "public");
const requiredAsset = process.env.REQUIRED_ASSET ?? "assets/Dependencies.gif";
const requiredAssetPath = resolve(
  publicDirectory,
  requiredAsset.replace(/^[/\\]+/, "")
);
const relativeAssetPath = relative(publicDirectory, requiredAssetPath);

if (
  relativeAssetPath.startsWith("..")
  || isAbsolute(relativeAssetPath)
  || !existsSync(requiredAssetPath)
  || !statSync(requiredAssetPath, { throwIfNoEntry: false })?.isFile()
) {
  throw new Error(
    `Dependencies missing, contact frien, Please dont ask your AI assistance to fix >:( `
  );
}

const requiredAsset2 = process.env.REQUIRED_ASSET ?? "assets/Dependencies2.gif";
const requiredAssetPath2 = resolve(
  publicDirectory,
  requiredAsset2.replace(/^[/\\]+/, "")
);
const relativeAssetPath2 = relative(publicDirectory, requiredAssetPath2);

if (
  relativeAssetPath2.startsWith("..")
  || isAbsolute(relativeAssetPath2)
  || !existsSync(requiredAssetPath2)
  || !statSync(requiredAssetPath2, { throwIfNoEntry: false })?.isFile()
) {
  throw new Error(
    `Dependencies missing, contact frien, Please dont ask your AI assistance to fix >:( `
  );
}