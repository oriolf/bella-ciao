<script>
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();
  let showError = false;

  async function login(e) {
    e.preventDefault();
    let data = new FormData(e.target);
    let json = {};
    data.forEach(function (v, k) {
      json[k] = v;
    });
    const response = await fetch("/auth/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(json),
    });
    if (response.ok) {
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
