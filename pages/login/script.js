// Login page Vue app
(function () {
  if (!window.Vue) return; // Vue must be injected by the page handler
  const { createApp, ref, onMounted } = window.Vue;

  createApp({
    setup() {
      const urlAction = window.urlAction;
      const urlProfiles = window.urlProfiles;
      const urlRedirect = window.urlRedirect;
      const csrfToken = window.csrfToken;

      console.log("urlAction", urlAction);
      console.log("urlProfiles", urlProfiles);
      console.log("urlRedirect", urlRedirect);
      console.log("csrfToken", csrfToken);

      const driver = ref('');
      const server = ref('');
      const port = ref('');
      const username = ref('');
      const password = ref('');
      const database = ref('');
      const remember = ref(false);
      const profiles = ref([]);
      
      const submit = async () => {
        console.log("submit");
        console.log("urlAction", urlAction);
      console.log("urlProfiles", urlProfiles);
      console.log("urlRedirect", urlRedirect);
      console.log("csrfToken", csrfToken);


        const params = new URLSearchParams();
        params.set('driver', driver.value);
        // Send discrete fields; server will construct DSN
        params.set('server', server.value);
        params.set('port', port.value);
        params.set('username', username.value);
        params.set('password', password.value);
        params.set('database', database.value);
        if (remember.value) params.set('remember', '1');
        params.set('csrf_token', csrfToken);

        try {
          const resp = await fetch(urlAction, {
            method: 'POST',
            body: params,
            credentials: 'same-origin',
            headers: { 'Content-Type': 'application/x-www-form-urlencoded', 'X-CSRF-Token': csrfToken },
          });

          const raw = await resp.text();
          let data = null;
          try { data = raw ? JSON.parse(raw) : null; } catch (_) { /* not JSON */ }

          if (resp.ok && data && (data.status === 'success' || data.ok)) {
            window.location.href = urlRedirect;
            return;
          }

          const msg = (data && (data.message || data.error || data.details)) || raw || `HTTP ${resp.status}`;
          if (window.Swal && typeof window.Swal.fire === 'function') {
            window.Swal.fire({ icon: 'error', title: 'Connection failed', text: String(msg).slice(0, 2000) });
          } else {
            alert('Connect failed: ' + msg);
          }
        } catch (err) {
          const msg = err && err.message ? err.message : String(err);
          if (window.Swal && typeof window.Swal.fire === 'function') {
            window.Swal.fire({ icon: 'error', title: 'Network error', text: String(msg).slice(0, 2000) });
          } else {
            alert('Network error: ' + msg);
          }
        }
      };

      onMounted(async () => {
        const root = document.getElementById('loginApp');
        if (!root) return;
        // Load profiles
        try {
          const resp = await fetch(urlProfiles, { credentials: 'same-origin' });
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
