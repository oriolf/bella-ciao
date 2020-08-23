<script>
  import { createEventDispatcher } from "svelte";
  import { submitForm } from "../util.js";

  const dispatch = createEventDispatcher();
  let showError = false;

  async function login(e) {
    const res = await submitForm(e, "/api/auth/login");
    if (res.ok) {
      dispatch("loggedin", true);
    } else {
      showError = true;
    }
  }
</script>

<div class="card">
  <div class="card-header">Log in</div>
  <div class="card-body">
    <form on:submit={login}>
      <div class="form-group">
        <label for="unique_id">User</label>
        <input
          class="form-control"
          id="unique_id"
          name="unique_id"
          aria-describedby="uniqueIdentifierHelp"
          placeholder="Enter user (identification number: DNI, NIE, etc.)" />
      </div>
      <div class="form-group">
        <label for="password">Password</label>
        <input
          type="password"
          class="form-control"
          id="password"
          name="password"
          placeholder="Password" />
      </div>
      <button type="submit" class="btn btn-primary">Submit</button>
    </form>
  </div>
  {#if showError}
    <div class="card-footer">
      <div class="alert alert-danger" role="alert" style="margin-bottom: 0;">
        The user or password is invalid
      </div>
    </div>
  {/if}
</div>
