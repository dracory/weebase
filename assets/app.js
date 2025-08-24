// Minimal embedded JS for WeeBase scaffold
console.log("WeeBase assets loaded");

(function () {
  const form = document.getElementById("adminerConnectForm");
  if (!form) return;
  form.addEventListener("submit", function (e) {
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
          // Example: host=localhost user=me password=secret dbname=mydb port=5432 sslmode=disable
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
          // Example: user:pass@tcp(localhost:3306)/dbname?parseTime=true
          const auth = user || pass ? `${user}:${pass}` : user;
          const dbpart = db ? `/${db}` : "";
          dsn = `${auth}@tcp(${hostPort})${dbpart}?parseTime=true`;
          break;
        }
        case "sqlite":
        case "sqlite3": {
          // Database field is the file path (e.g., :memory: or ./data.db)
          dsn = db || ":memory:";
          break;
        }
        case "sqlserver":
        case "mssql": {
          // Example: sqlserver://user:pass@localhost:1433?database=dbname
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
    } catch (err) {
      console.warn("failed constructing dsn:", err);
    }
  });
})();
