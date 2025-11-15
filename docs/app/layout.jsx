import { Footer, Layout, Navbar } from "nextra-theme-docs";
import { Head } from "nextra/components";
import { getPageMap } from "nextra/page-map";
import "nextra-theme-docs/style.css";
import VersionSelector from "./components/VersionSelector";

export const metadata = {
  // Define your metadata here
  // For more information on metadata API, see: https://nextjs.org/docs/app/building-your-application/optimizing/metadata
};

// const banner = <Banner storageKey="some-key">Drift 1.0 is released 🎉</Banner>;
const banner = undefined;
const navbar = (
  <Navbar>
    <div style={{ order: -1, display: "flex", alignItems: "center" }}>
      <div style={{ marginRight: 24 }}>
        <b>Drift</b>
      </div>
      <VersionSelector />
    </div>
  </Navbar>
);
const footer = (
  <Footer>Apache License 2.0 {new Date().getFullYear()} © Driftee.</Footer>
);

export default async function RootLayout({ children }) {
  return (
    <html
      // Not required, but good for SEO
      lang="en"
      // Required to be set
      dir="ltr"
      // Suggested by `next-themes` package https://github.com/pacocoursey/next-themes#with-app
      suppressHydrationWarning
    >
      <Head
      // ... Your additional head options
      >
        {/* Your additional tags should be passed as `children` of `<Head>` element */}
      </Head>
      <body>
        <Layout
          banner={banner}
          navbar={navbar}
          pageMap={await getPageMap()}
          docsRepositoryBase="https://github.com/shuding/nextra/tree/main/docs"
          footer={footer}
          // ... Your additional layout options
        >
          {children}
        </Layout>
      </body>
    </html>
  );
}
