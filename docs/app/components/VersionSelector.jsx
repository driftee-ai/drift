"use client";

import { useEffect, useState } from "react";

const VersionSelector = ({ version }) => {
  const [versions, setVersions] = useState([version]);
  const [currentVersion, setCurrentVersion] = useState(version);
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    console.log("in use effect");
    if (process.env.NODE_ENV === "production") {
      fetch("/versions.json")
        .then((res) => {
          if (!res.ok) {
            console.error(`Failed to fetch versions.json: ${res.status}`);
            return [version];
          }
          return res.json();
        })
        .then((data) => {
          if (Array.isArray(data) && data.length > 0) {
            setVersions(data);
            const pathParts = window.location.pathname.split("/");
            if (
              pathParts.length > 1 &&
              (data.includes(pathParts[1]) || pathParts[1] === "latest")
            ) {
              setCurrentVersion(pathParts[1]);
            } else {
              setCurrentVersion("latest");
            }
          } else {
            setVersions([version]);
            setCurrentVersion(version);
          }
        })
        .catch((error) => {
          console.error("Error fetching or parsing versions.json:", error);
          setVersions([version]);
          setCurrentVersion(version);
        })
        .finally(() => {
          setIsLoaded(true);
        });
    } else {
      // In development, provide dummy versions for testing styling and interaction
      const dummyVersions = ["1.0", "2.0", "latest"];
      setVersions(dummyVersions);

      const pathParts = window.location.pathname.split("/");
      // Try to set currentVersion based on path, or default to 'latest'
      // If the current path segment is one of the dummy versions, use it.
      // Otherwise, default to 'latest' for display.
      if (
        pathParts.length > 1 &&
        (dummyVersions.includes(pathParts[1]) || pathParts[1] === "latest")
      ) {
        setCurrentVersion(pathParts[1]);
      } else {
        setCurrentVersion("latest");
      }
      setIsLoaded(true);
    }
  }, []);

  const handleVersionChange = (e) => {
    const newVersion = e.target.value;
    const pathParts = window.location.pathname.split("/");

    if (process.env.NODE_ENV !== "production") {
      // In development, for dummy versions, update the URL to reflect the selected version.
      // The content will remain the same as there are no separate docs for dummy versions.
      const currentPathSegment = pathParts.length > 1 ? pathParts[1] : "";
      const restOfPath =
        pathParts.length > 2 ? pathParts.slice(2).join("/") : "";

      if (
        currentPathSegment &&
        (versions.includes(currentPathSegment) ||
          currentPathSegment === "latest")
      ) {
        console.log("case a");
        window.location.pathname = `/${newVersion}/${restOfPath}`;
      } else {
        console.log("case b");
        window.location.pathname = `/${newVersion}/${pathParts.slice(1).join("/")}`;
      }
      return;
    }

    // Production logic
    if (
      pathParts.length > 2 &&
      (versions.includes(pathParts[1]) || pathParts[1] === "latest")
    ) {
      const restOfPath = pathParts.slice(2).join("/");
      console.log("case c");
      window.location.pathname = `/${newVersion}/${restOfPath}`;
    } else {
      console.log("case d");
      window.location.pathname = `/${newVersion}/`;
    }
  };

  return (
    <select
      value={currentVersion}
      onChange={handleVersionChange}
      disabled={!isLoaded && process.env.NODE_ENV === "production"} // Only disabled if in production and not loaded
    >
      {versions.map((v) => (
        <option key={v} value={v}>
          {v}
        </option>
      ))}
    </select>
  );
};

export default VersionSelector;
