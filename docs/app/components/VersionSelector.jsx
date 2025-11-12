"use client";

import { useEffect, useState } from "react";
import { useRouter, usePathname } from "next/navigation";

const VersionSelector = ({ version }) => {
  const router = useRouter();
  const pathname = usePathname();
  const { basePath } = router;

  const [versions, setVersions] = useState([version]);
  const [currentVersion, setCurrentVersion] = useState(version);
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    if (process.env.NODE_ENV === "production") {
      fetch("/versions.json")
        .then((res) => (res.ok ? res.json() : []))
        .then((data) => {
          if (Array.isArray(data) && data.length > 0) {
            setVersions(data);
          }
        })
        .catch((error) => {
          console.error("Error fetching or parsing versions.json:", error);
        })
        .finally(() => {
          const versionFromPath =
            basePath && basePath.startsWith("/")
              ? basePath.substring(1)
              : basePath;
          if (versionFromPath) {
            setCurrentVersion(versionFromPath);
          }
          setIsLoaded(true);
        });
    } else {
      const dummyVersions = ["latest", "v0.2.0", "v0.1.0"];
      setVersions(dummyVersions);
      const versionFromPath =
        basePath && basePath.startsWith("/")
          ? basePath.substring(1)
          : basePath;
      if (dummyVersions.includes(versionFromPath)) {
        setCurrentVersion(versionFromPath);
      } else {
        setCurrentVersion("latest");
      }
      setIsLoaded(true);
    }
  }, [basePath, version]);

  const handleVersionChange = (e) => {
    const newVersion = e.target.value;
    window.location.href = `/${newVersion}${pathname}`;
  };

  return (
    <select
      value={currentVersion}
      onChange={handleVersionChange}
      disabled={!isLoaded && process.env.NODE_ENV === "production"}
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
