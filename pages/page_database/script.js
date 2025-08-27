// Database browser Vue app
(function () {
  if (!window.Vue) return; // Vue must be injected by the page handler
  const { createApp, ref, onMounted } = window.Vue;

  createApp({
    setup() {
      // State
      const isLoading = ref(true);
      const error = ref('');
      const databases = ref([]);
      const selectedDatabase = ref(null);
      const tables = ref([]);
      const tablesLoading = ref(false);

      // Format file size
      const formatSize = (bytes) => {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
      };

      // Format date
      const formatDate = (dateString) => {
        if (!dateString) return 'N/A';
        const date = new Date(dateString);
        return date.toLocaleString();
      };

      // Fetch databases from the server
      const fetchDatabases = async () => {
        isLoading.value = true;
        error.value = '';
        
        try {
          const response = await fetch('/api/databases', {
            credentials: 'same-origin',
            headers: {
              'Accept': 'application/json',
              'X-Requested-With': 'XMLHttpRequest'
            }
          });
          
          if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
          }
          
          const data = await response.json();
          if (data.status === 'success' && Array.isArray(data.data)) {
            databases.value = data.data;
          } else {
            throw new Error(data.message || 'Failed to load databases');
          }
        } catch (err) {
          console.error('Error fetching databases:', err);
          error.value = err.message || 'Failed to load databases. Please try again.';
        } finally {
          isLoading.value = false;
        }
      };

      // Fetch tables for a database
      const fetchTables = async (dbName) => {
        if (!dbName) return;
        
        tablesLoading.value = true;
        error.value = '';
        
        try {
          const response = await fetch(`/api/tables?db=${encodeURIComponent(dbName)}`, {
            credentials: 'same-origin',
            headers: {
              'Accept': 'application/json',
              'X-Requested-With': 'XMLHttpRequest'
            }
          });
          
          if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
          }
          
          const data = await response.json();
          if (data.status === 'success' && Array.isArray(data.data)) {
            tables.value = data.data;
          } else {
            throw new Error(data.message || 'Failed to load tables');
          }
        } catch (err) {
          console.error('Error fetching tables:', err);
          error.value = err.message || 'Failed to load tables. Please try again.';
        } finally {
          tablesLoading.value = false;
        }
      };

      // Select a database to view its tables
      const selectDatabase = (db) => {
        selectedDatabase.value = db;
        tables.value = [];
        fetchTables(db.name);
      };

      // Select a table to view its data
      const selectTable = (table) => {
        // Navigate to the table view
        window.location.href = `/table/${encodeURIComponent(selectedDatabase.value.name)}/${encodeURIComponent(table.name)}`;
      };

      // Refresh the database list
      const refresh = () => {
        fetchDatabases();
      };

      // Disconnect from the current database
      const disconnect = () => {
        window.location.href = '/logout';
      };

      // Initialize
      onMounted(() => {
        fetchDatabases();
      });

      // Expose to template
      return {
        isLoading,
        error,
        databases,
        selectedDatabase,
        tables,
        tablesLoading,
        formatSize,
        formatDate,
        selectDatabase,
        selectTable,
        refresh,
        disconnect
      };
    },
  }).mount('.database-browser');
})();
