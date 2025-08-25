// Login page Vue app
(function () {
  if (!window.Vue) return; // Vue must be injected by the page handler
  const { createApp, ref, onMounted } = window.Vue;

  createApp({
    setup() {
      const driver = ref('');
      const server = ref('');
      const port = ref('');
      const username = ref('');
      const password = ref('');
      const database = ref('');
      const remember = ref(false);
      const profiles = ref([]);
      let basePath = '';
      let actionParam = '';
      let csrfToken = '';

      const buildDSN = () => {
        const hostPort = server.value + (port.value ? ':' + port.value : '');
        switch ((driver.value || '').toLowerCase()) {
          case 'postgres':
          case 'pg':
          case 'postgresql': {
            const parts = [];
            if (server.value) parts.push(`host=${server.value}`);
            if (username.value) parts.push(`user=${username.value}`);
            if (password.value) parts.push(`password=${password.value}`);
            if (database.value) parts.push(`dbname=${database.value}`);
            if (port.value) parts.push(`port=${port.value}`);
            parts.push('sslmode=disable');
            return parts.join(' ');
          }
          case 'mysql':
          case 'mariadb': {
            const auth = username.value || password.value ? `${username.value}:${password.value}` : username.value;
            const dbpart = database.value ? `/${database.value}` : '';
            return `${auth}@tcp(${hostPort})${dbpart}?parseTime=true`;
          }
          case 'sqlite':
          case 'sqlite3': {
            return database.value || ':memory:';
          }
          case 'sqlserver':
          case 'mssql': {
            const auth = username.value || password.value ? `${encodeURIComponent(username.value)}:${encodeURIComponent(password.value)}@` : '';
            const qp = database.value ? `?database=${encodeURIComponent(database.value)}` : '';
            return `sqlserver://${auth}${hostPort}${qp}`;
          }
          default:
            return '';
        }
      };

      const submit = async () => {
        const actionURL = `${basePath}?${actionParam}=connect`;
        const params = new URLSearchParams();
        params.set('driver', driver.value);
        params.set('dsn', buildDSN());
        if (remember.value) params.set('remember', '1');
        params.set('csrf_token', csrfToken);
        const resp = await fetch(actionURL, {
          method: 'POST',
          body: params,
          credentials: 'same-origin',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded', 'X-CSRF-Token': csrfToken },
        });
        const data = await resp.json().catch(() => null);
        if (resp.ok && data && (data.status === 'success' || data.ok)) {
          window.location.href = basePath;
        } else {
          const msg = (data && (data.message || data.error)) || `HTTP ${resp.status}`;
          alert('Connect failed: ' + msg);
        }
      };

      onMounted(async () => {
        const root = document.getElementById('loginApp');
        if (!root) return;
        basePath = root.dataset.basepath || '';
        actionParam = root.dataset.actionparam || 'action';
        csrfToken = root.dataset.csrf || '';
        // Load profiles
        try {
          const url = `${basePath}?${actionParam}=profiles`;
          const resp = await fetch(url, { credentials: 'same-origin' });
          const data = await resp.json().catch(() => null);
          if (resp.ok && data && data.data && Array.isArray(data.data.profiles)) {
            profiles.value = data.data.profiles;
          }
        } catch (_) { /* ignore */ }
      });

      return { driver, server, port, username, password, database, remember, profiles, submit };
    },
  }).mount('#loginApp');
})();
