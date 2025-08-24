// Minimal embedded JS for WeeBase scaffold
console.log("WeeBase assets loaded");

(function () {
  const form = document.getElementById("adminerConnectForm");
  if (!form) return;
  form.addEventListener("submit", async function (e) {
    e.preventDefault();
    try {
      const driver = form.querySelector('[name="driver"]').value.trim();
      const server = (form.querySelector('[name="server"]').value || "").trim();
      const port = (form.querySelector('[name="port"]').value || "").trim();
      const user = (form.querySelector('[name="username"]').value || "").trim();
      const pass = (form.querySelector('[name="password"]').value || "").trim();
      const db = (form.querySelector('[name="database"]').value || "").trim();
      let dsn = "";
      const hostPort = server + (port ? ":" + port : "");
      switch (driver.toLowerCase()) {
        case "postgres":
        case "pg":
        case "postgresql": {
          const parts = [];
          if (server) parts.push(`host=${server}`);
          if (user) parts.push(`user=${user}`);
          if (pass) parts.push(`password=${pass}`);
          if (db) parts.push(`dbname=${db}`);
          if (port) parts.push(`port=${port}`);
          parts.push("sslmode=disable");
          dsn = parts.join(" ");
          break;
        }
        case "mysql":
        case "mariadb": {
          const auth = user || pass ? `${user}:${pass}` : user;
          const dbpart = db ? `/${db}` : "";
          dsn = `${auth}@tcp(${hostPort})${dbpart}?parseTime=true`;
          break;
        }
        case "sqlite":
        case "sqlite3": {
          dsn = db || ":memory:";
          break;
        }
        case "sqlserver":
        case "mssql": {
          const auth = user || pass ? `${encodeURIComponent(user)}:${encodeURIComponent(pass)}@` : "";
          const qp = db ? `?database=${encodeURIComponent(db)}` : "";
          dsn = `sqlserver://${auth}${hostPort}${qp}`;
          break;
        }
        default:
          // leave empty; server will error with unsupported driver
          break;
      }
      form.querySelector('[name="dsn"]').value = dsn;

      const fd = new FormData(form);
      const resp = await fetch(form.action, {
        method: "POST",
        body: fd,
        credentials: "same-origin",
      });
      const data = await resp.json().catch(() => null);
      if (resp.ok && data && (data.status === "success" || data.ok)) {
        // Redirect to home
        window.location.href = form.action;
      } else {
        const msg = (data && (data.message || data.error)) || `HTTP ${resp.status}`;
        alert("Connect failed: " + msg);
      }
    } catch (err) {
      alert("Unexpected error: " + (err && err.message ? err.message : err));
    }
  });
})();
