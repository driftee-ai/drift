"use client";

import { useEffect, useState } from "react";

const VersionSelector = ({ version }) => {
  const [versions, setVersions] = useState([version]);
  const [currentVersion, setCurrentVersion] = useState("");

  useEffect(() => {
    fetch("/versions.json")
      .then((res) => res.json())
      .then((data) => {
        console.log("data.json", data);
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
      });
  }, []);

  const handleVersionChange = (e) => {
    const newVersion = e.target.value;
    const pathParts = window.location.pathname.split("/");
    if (
      pathParts.length > 2 &&
      (versions.includes(pathParts[1]) || pathParts[1] === "latest")
    ) {
      const restOfPath = pathParts.slice(2).join("/");
      window.location.pathname = `/${newVersion}/${restOfPath}`;
    } else {
      window.location.pathname = `/${newVersion}/`;
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

export default VersionSelector;
