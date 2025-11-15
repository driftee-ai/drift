"use client";

import { useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { uniq } from "lodash";

const VersionSelector = () => {
  const pathname = usePathname();
  const versionFromPath = pathname.split("/").filter(Boolean)[0] || "latest";
  console.log("pathname", pathname);
  console.log("versionFromPath", versionFromPath);

  const [versions, setVersions] = useState([versionFromPath]);
  const [currentVersion, setCurrentVersion] = useState(versionFromPath);

  useEffect(() => {
    const fetchVersions = async () => {
      try {
        const url = `https://docs.driftee.ai/versions.json?t=${new Date().getTime()}`;
        const res = await fetch(url);
        if (res.ok) {
          const data = await res.json();
          // Use a functional update to avoid race conditions with state
          setVersions((prev) => uniq([...prev, ...data, "latest"]));
        }
      } catch (error) {
        console.error("Failed to fetch versions:", error);
      }
    };

    fetchVersions();
  }, []); // Fetch only once

  const handleVersionChange = (e) => {
    const newVersion = e.target.value;
    window.location.href = `/${newVersion}/`;
  };

  return (
    <select value={currentVersion} onChange={handleVersionChange}>
      {versions.map((v) => (
        <option key={v} value={v}>
          {v}
        </option>
      ))}
    </select>
  );
};

export default VersionSelector;
