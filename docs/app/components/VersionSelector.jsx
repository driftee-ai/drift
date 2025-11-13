"use client";

import { useEffect, useState } from "react";
import { usePathname } from "next/navigation";

const fetchVersions = async () => {
  try {
    const res = await fetch("/versions.json");
    const data = res.ok ? await res.json() : [];
    console.log("data", data);
    return data;
  } catch (error) {
    console.error("Error fetching or parsing versions.json:", error);
    return ["latest", "v0.2.0", "v0.1.0"];
  }
};

const VersionSelector = ({ version }) => {
  const pathname = usePathname();

  const [versions, setVersions] = useState([...new Set([version, "latest"])]);
  const [currentVersion, setCurrentVersion] = useState(version);
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(async () => {
    const newVersions = await fetchVersions();
    setVersions([...versions, ...newVersions]);
    setIsLoaded(true);
  }, []);

  const handleVersionChange = (e) => {
    const newVersion = e.target.value;
    if (process.env.NODE_ENV === "production") {
      window.location.href = `/${newVersion}`;
    } else {
      setCurrentVersion(newVersion);
      console.log("setting new version in dev", newVersion);
    }
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
