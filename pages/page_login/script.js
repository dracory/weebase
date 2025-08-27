// Login page Vue app
(function () {
  if (!window.Vue) return; // Vue must be injected by the page handler
  const { createApp, ref, computed, watch } = window.Vue;

  createApp({
    setup() {
      // Configuration from server
      const urlAction = window.urlAction || '';
      const urlProfiles = window.urlProfiles || '';
      const urlRedirect = window.urlRedirect || '';
      const csrfToken = window.csrfToken || '';

      // Form state
      const driver = ref('sqlite');
      const server = ref('localhost');
      const port = ref('');
      const username = ref('');
      const password = ref('');
      const database = ref('');
      const remember = ref(false);
      const isLoading = ref(false);
      const error = ref('');
      const profiles = ref([]);

      // Set default ports based on selected driver
      const defaultPorts = {
        mysql: '3306',
        postgres: '5432',
        sqlserver: '1433',
        sqlite: ''
      };

      // Watch for driver changes to update port
      watch(driver, (newDriver) => {
        if (newDriver && defaultPorts[newDriver] && !port.value) {
          port.value = defaultPorts[newDriver];
        }
        // Clear database field when switching to SQLite
        if (newDriver === 'sqlite') {
          database.value = '';
        }
      }, { immediate: true });

      // Form validation
      const isFormValid = computed(() => {
        if (driver.value === 'sqlite') {
          return database.value.trim() !== '';
        }
        return (
          server.value.trim() !== '' &&
          username.value.trim() !== ''
        );
      });

      // Handle form submission
      const submit = async () => {
        if (!isFormValid.value) {
          showError('Please fill in all required fields');
          return;
        }

        isLoading.value = true;
        error.value = '';

        const params = new URLSearchParams();
        params.set('driver', driver.value);
        
        // Only include non-SQLite fields when not using SQLite
        if (driver.value !== 'sqlite') {
          params.set('server', server.value.trim());
          if (port.value) params.set('port', port.value);
          params.set('username', username.value.trim());
          if (password.value) params.set('password', password.value);
          if (database.value) params.set('database', database.value.trim());
        } else {
          // For SQLite, only the database path is needed
          params.set('database', database.value.trim());
        }
        
        if (remember.value) params.set('remember', '1');
        if (csrfToken) params.set('csrf_token', csrfToken);

        try {
          const response = await fetch(urlAction, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/x-www-form-urlencoded',
              'X-Requested-With': 'XMLHttpRequest',
              ...(csrfToken && { 'X-CSRF-Token': csrfToken })
            },
            credentials: 'same-origin',
            body: params
          });

          const data = await response.json().catch(() => ({}));
          
          if (response.ok && (data.status === 'success' || data.ok)) {
            // Successful login, redirect
            window.location.href = urlRedirect;
            return;
          }

          // Handle error response
          const errorMessage = data?.message || data?.error || 'Connection failed';
          showError(errorMessage);
        } catch (err) {
          console.error('Login error:', err);
          showError('Network error. Please check your connection and try again.');
        } finally {
          isLoading.value = false;
        }
      };

      // Load saved profiles if available
      const loadProfiles = async () => {
        try {
          const response = await fetch(urlProfiles, {
            credentials: 'same-origin',
            headers: {
              'Accept': 'application/json',
              'X-Requested-With': 'XMLHttpRequest'
            }
          });
          
          if (response.ok) {
            const data = await response.json();
            if (data.data?.profiles?.length) {
              profiles.value = data.data.profiles;
            }
          }
        } catch (err) {
          console.error('Failed to load profiles:', err);
        }
      };

      // Apply profile settings
      const applyProfile = (profile) => {
        if (!profile) return;
        
        driver.value = profile.driver || 'sqlite';
        server.value = profile.server || '';
        port.value = profile.port || defaultPorts[driver.value] || '';
        username.value = profile.username || '';
        database.value = profile.database || '';
      };

      // Show error message
      const showError = (message) => {
        error.value = message;
        // Auto-hide error after 5 seconds
        setTimeout(() => { error.value = ''; }, 5000);
      };

      // Initialize
      loadProfiles();

      // Expose to template
      return {
        driver,
        server,
        port,
        username,
        password,
        database,
        remember,
        isLoading,
        error,
        profiles,
        isFormValid,
        submit,
        applyProfile
      };
    },
  }).mount('#loginApp');
})();
