// Sidebar Vue component
export default {
  template: `
    <div id="sidebar-app">
      <ul class="nav flex-column">
        <li v-if="loading" class="nav-item">
          <span class="nav-link text-muted">loading...</span>
        </li>
        <li v-else-if="error" class="nav-item">
          <span class="text-red-600">{{ error }}</span>
        </li>
        <li v-else-if="tables.length === 0" class="nav-item">
          <span class="nav-link text-muted">No tables</span>
        </li>
        <li v-else v-for="table in tables" :key="table" class="nav-item">
          <a 
            :href="getTableUrl(table)" 
            class="nav-link text-dark hover:underline"
            :title="'Select ' + table"
          >
            {{ table }}
          </a>
        </li>
      </ul>
    </div>
  `,
  data() {
    return {
      tables: [],
      loading: true,
      error: ''
    }
  },
  async mounted() {
    await this.loadTables();
  },
  methods: {
    async loadTables() {
      try {
        const url = window.urlListTables || '';
        const response = await fetch(url, { credentials: 'same-origin' });
        const data = await response.json();
        this.tables = (data?.data?.tables && Array.isArray(data.data.tables)) 
          ? data.data.tables 
          : [];
      } catch (err) {
        this.error = err?.message || String(err);
      } finally {
        this.loading = false;
      }
    },
    getTableUrl(table) {
      const base = window.urlTable || window.urlBrowseBase || '';
      const separator = base.includes('?') ? '&' : '?';
      return `${base}${separator}table=${encodeURIComponent(table)}`;
    }
  }
};
