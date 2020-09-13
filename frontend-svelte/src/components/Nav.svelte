<script lang="ts">
  import { user } from "../store";

  export let segment: string;
  let links = [
    { segment: undefined, href: ".", name: "Home" },
    { segment: "candidates", href: "candidates", name: "Candidates" },
    { segment: "faq", href: "faq", name: "FAQ" },
  ];

  async function logout() {
    await fetch("/api/auth/logout");
    user.set(null);
    window.location.href = "/";
  }
</script>

<style>
  nav.navbar {
    margin-bottom: 10px;
  }
</style>

<nav class="navbar navbar-expand-lg sticky-top navbar-dark bg-primary">
  <span class="navbar-brand">Bella Ciao</span>
  <button
    class="navbar-toggler"
    type="button"
    data-toggle="collapse"
    data-target="#navbar"
    aria-controls="navbar"
    aria-expanded="false"
    aria-label="Toggle navigation">
    <span class="navbar-toggler-icon" />
  </button>
  <div class="collapse navbar-collapse" id="navbar">
    <ul class="navbar-nav mr-auto">
      {#each links as link}
        <li class="nav-item" class:active={segment === link.segment}>
          <a
            class="nav-link"
            aria-current={segment === link.segment ? 'page' : undefined}
            href={link.href}>
            {link.name}
            {#if segment === link.segment}
              <span class="sr-only">(current)</span>
            {/if}
          </a>
        </li>
      {/each}
    </ul>
    {#if $user}
      <span class="navbar-text">
        <a class="nav-link" href="." on:click={logout}>Log out</a>
      </span>
    {/if}
  </div>
</nav>
