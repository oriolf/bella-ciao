<script>
  import Error from "../routes/_error.svelte";
  import LoginForm from "./LoginForm.svelte";

  async function whoami() {
    let res = await fetch("/users/whoami");
    if (!res.ok) {
      throw new Error("Not logged in");
    }
    return res.json();
  }
  let promise = whoami();
  function getUser() {
      promise = whoami();
  }
</script>

<svelte:head>
  <title>Bella Ciao</title>
</svelte:head>

{#await promise}
  <p>...waiting</p>
{:then user}
  <p>Im am {user.role}</p>
{:catch _}
  <LoginForm on:loggedin="{getUser}"/>
{/await}
