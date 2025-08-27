// Table viewer Vue app
(function () {
  if (!window.Vue) return; // Vue must be injected by the page handler
  const { createApp, ref, onMounted, computed } = window.Vue;

  createApp({
    setup() {
      // State
      const isLoading = ref(true);
      const error = ref('');
      const tableData = ref([]);
      const columns = ref([]);
      const databaseName = ref('');
      const tableName = ref('');
      const basePath = ref('/');
      
      // Pagination
      const currentPage = ref(1);
      const pageSize = ref(25);
      const totalRows = ref(0);
      
      // Sorting
      const sortColumn = ref('');
      const sortDirection = ref('asc');
      
      // Search
      const searchQuery = ref('');
      
      // Format JSON for display
      const formatJson = (value) => {
        if (!value) return '';
        try {
          return JSON.stringify(JSON.parse(value), null, 2);
        } catch (e) {
          return value;
        }
      };
      
      // Format cell value for display
      const formatCellValue = (value) => {
        if (value === null || value === undefined) {
          return '<span class="null">NULL</span>';
        }
        if (typeof value === 'boolean') {
          return value ? '✓' : '✗';
        }
        return value;
      };
      
      // Fetch table data from the server
      const fetchTableData = async () => {
        isLoading.value = true;
        error.value = '';
        
        try {
          const params = new URLSearchParams({
            page: currentPage.value,
            pageSize: pageSize.value,
            sort: sortColumn.value,
            order: sortDirection.value,
            q: searchQuery.value
          });
          
          const url = `${window.appConfig.api.tables}?${params.toString()}`;
          const response = await fetch(url, {
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
          if (data.status === 'success') {
            tableData.value = data.data.rows || [];
            columns.value = data.data.columns || [];
            totalRows.value = data.data.total || 0;
            
            if (!databaseName.value && data.data.database) {
              databaseName.value = data.data.database;
            }
            if (!tableName.value && data.data.table) {
              tableName.value = data.data.table;
            }
          } else {
            throw new Error(data.message || 'Failed to load table data');
          }
        } catch (err) {
          console.error('Error fetching table data:', err);
          error.value = err.message || 'Failed to load table data. Please try again.';
        } finally {
          isLoading.value = false;
        }
      };
      
      // Handle sort column click
      const sortBy = (column) => {
        if (sortColumn.value === column) {
          // Toggle sort direction if same column
          sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc';
        } else {
          // New column, default to ascending
          sortColumn.value = column;
          sortDirection.value = 'asc';
        }
        fetchTableData();
      };
      
      // Pagination handlers
      const nextPage = () => {
        if (currentPage.value * pageSize.value < totalRows.value) {
          currentPage.value++;
          fetchTableData();
        }
      };
      
      const prevPage = () => {
        if (currentPage.value > 1) {
          currentPage.value--;
          fetchTableData();
        }
      };
      
      const changePageSize = (event) => {
        pageSize.value = parseInt(event.target.value, 10);
        currentPage.value = 1; // Reset to first page
        fetchTableData();
      };
      
      // Refresh table data
      const refreshTable = () => {
        fetchTableData();
      };
      
      // Initialize
      onMounted(() => {
        // Get database and table names from URL
        const pathSegments = window.location.pathname.split('/').filter(Boolean);
        if (pathSegments.length >= 2) {
          databaseName.value = decodeURIComponent(pathSegments[pathSegments.length - 2]);
          tableName.value = decodeURIComponent(pathSegments[pathSegments.length - 1]);
        }
        
        // Get base path (everything before /table/db/table)
        const basePathMatch = window.location.pathname.match(/^(.*?)\/table\/.*$/);
        if (basePathMatch && basePathMatch[1]) {
          basePath.value = basePathMatch[1] + '/';
        }
        
        // Initial data load
        fetchTableData();
      });

      // Expose to template
      return {
        isLoading,
        error,
        tableData,
        columns,
        databaseName,
        tableName,
        basePath,
        currentPage,
        pageSize,
        totalRows,
        sortColumn,
        sortDirection,
        searchQuery,
        formatJson,
        formatCellValue,
        sortBy,
        nextPage,
        prevPage,
        changePageSize,
        refreshTable
      };
    },
  }).mount('.table-viewer');
})();
