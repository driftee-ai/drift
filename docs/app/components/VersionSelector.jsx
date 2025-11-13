"use client";

import { useEffect, useState } from "react";
import { uniq } from "lodash";

const fetchVersions = async () => {
  if (process.env.NODE_ENV === "production") {
    const url = `${window.location.origin}/versions.json?t=${new Date().getTime()}`;
    const res = await fetch(url);
    const data = res.ok ? await res.json() : [];
    console.log("data in prod", data);
    return data;
  } else {
    return ["latest", "v0.2.0", "v0.1.0"];
  }
};

const VersionSelector = ({ version }) => {
  const [versions, setVersions] = useState(uniq([version, "latest"]));
  const [currentVersion, setCurrentVersion] = useState(version);
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    const pathSegments = window.location.pathname.split("/").filter(Boolean);
    const versionFromPath = pathSegments[0] || "latest";
    console.log("Version from prop (build time):", version);
    console.log("Version from URL (run time):", versionFromPath);

    const fetchData = async () => {
      const newVersions = await fetchVersions();
      setVersions((prev) => uniq([...prev, ...newVersions]));
      setIsLoaded(true);
    };
    fetchData();
  }, [version]);

  const handleVersionChange = (e) => {
    const newVersion = e.target.value;
    if (process.env.NODE_ENV === "production") {
      window.location.href = `/${newVersion}/`;
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
