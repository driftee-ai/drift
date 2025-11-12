import nextra from "nextra";

const isProd = process.env.NODE_ENV === "production";
const basePath = process.env.BASE_PATH || "";

// Set up Nextra with its configuration
const withNextra = nextra({
  // ... Add Nextra-specific options here
});

// Export the final Next.js config with Nextra included
export default withNextra({
  output: "export",
  basePath: basePath,
  trailingSlash: true,
  // ... Add regular Next.js options here
  async headers() {
    return [
      {
        source: "/versions.json",
        headers: [
          {
            key: "Cache-Control",
            value: "no-cache",
          },
        ],
      },
    ];
  },
});
