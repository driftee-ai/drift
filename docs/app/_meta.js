import { GitHubIcon } from "nextra/icons";

export default {
  index: "My Homepage",
  // You can use JSX elements to change the look of titles in the sidebar, e.g. insert icons
  about: {
    // Alternatively, you can set title with `title` property
    title: "About Us",
    // ... and provide extra configurations
  },
};

// Custom component for italicized text
function Italic({ children, ...props }) {
  return <i {...props}>{children}</i>;
}
