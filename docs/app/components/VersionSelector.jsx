"use client";

import { useEffect, useState } from "react";
import { uniq } from "lodash";

const fetchVersions = async () => {
  if (process.env.NODE_ENV === "production") {
    const res = await fetch("/versions.json");
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
    const fetchData = async () => {
      const newVersions = await fetchVersions();
      setVersions(uniq([...versions, ...newVersions]));
      setIsLoaded(true);
    };
    fetchData();
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
