<script>
  import UsersPagination from "./UsersPagination.svelte";
</script>

<h2>Users pending validation</h2>

<UsersPagination
  error="Could not get users pending validation"
  url="/api/users/unvalidated/get" />

<h2>Validated users</h2>

<UsersPagination
  error="Could not get validated users"
  url="/api/users/validated/get" />
