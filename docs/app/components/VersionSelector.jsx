"use client";

import { useEffect, useState } from "react";
import { uniq } from "lodash";

const fetchVersions = async () => {
  if (process.env.NODE_ENV === "production") {
    const url = `${window.location.origin}/versions.json?t=${new Date().getTime()}`;
    const res = await fetch(url);
    const data = res.ok ? await res.json() : [];
    return data;
  } else {
    return ["latest", "v0.2.0", "v0.1.0"];
  }
};

const VersionSelector = () => {
  const [versions, setVersions] = useState([]);
  const [currentVersion, setCurrentVersion] = useState();
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    const pathSegments = window.location.pathname.split("/").filter(Boolean);
    const versionFromPath = pathSegments[0] || "latest";
    setCurrentVersion(versionFromPath);
    setVersions(uniq(["latest", versionFromPath]));

    const fetchData = async () => {
      const newVersions = await fetchVersions();
      setVersions((prev) => uniq([...prev, ...newVersions]));
      setIsLoaded(true);
    };
    fetchData();
  }, []);

  const handleVersionChange = (e) => {
    const newVersion = e.target.value;
    const currentPath = window.location.pathname.split("/")[1] || "latest";

    if (newVersion !== currentPath) {
      if (process.env.NODE_ENV === "production") {
        window.location.href = `/${newVersion}/`;
      } else {
        setCurrentVersion(newVersion);
        console.log("setting new version in dev", newVersion);
      }
    }
  };

  if (!currentVersion) {
    return <p>no currentVersion found</p>;
  }

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
