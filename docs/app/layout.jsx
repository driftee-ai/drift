
import { Footer, Layout, Navbar } from "nextra-theme-docs";
import { Banner, Head } from "nextra/components";
import { getPageMap } from "nextra/page-map";
import "nextra-theme-docs/style.css";
import { useEffect, useState } from "react";

export const metadata = {
  // Define your metadata here
  // For more information on metadata API, see: https://nextjs.org/docs/app/building-your-application/optimizing/metadata
};

const VersionSelector = () => {
  const [versions, setVersions] = useState([]);
  const [currentVersion, setCurrentVersion] = useState("");

  useEffect(() => {
    fetch("/versions.json")
      .then((res) => res.json())
      .then((data) => {
        setVersions(data);
        const pathParts = window.location.pathname.split("/");
        if (pathParts.length > 1 && data.includes(pathParts[1])) {
          setCurrentVersion(pathParts[1]);
        } else {
          setCurrentVersion("latest");
        }
      });
  }, []);

  const handleVersionChange = (e) => {
    const newVersion = e.target.value;
    const pathParts = window.location.pathname.split("/");
    if (pathParts.length > 2 && versions.includes(pathParts[1])) {
      window.location.pathname = `/${newVersion}/${pathParts.slice(2).join("/")}`;
    } else {
      window.location.pathname = `/${newVersion}`;
    }
  };

  return (
    <select value={currentVersion} onChange={handleVersionChange}>
      {versions.map((version) => (
        <option key={version} value={version}>
          {version}
        </option>
      ))}
    </select>
  );
};

const banner = <Banner storageKey="some-key">Nextra 4.0 is released 🎉</Banner>;
const navbar = (
  <Navbar
    logo={<b>Nextra</b>}
    extra={<VersionSelector />}
    // ... Your additional navbar options
  />
);
const footer = <Footer>MIT {new Date().getFullYear()} © Nextra.</Footer>;

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

