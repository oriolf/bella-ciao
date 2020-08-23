<script>
  import { goto } from "@sapper/app";
  // TODO only show logout if logged in

  export let segment;
  let links = [
    { segment: undefined, href: ".", name: "Home" },
    { segment: "faq", href: "faq", name: "FAQ" },
  ];

  async function logout() {
    await fetch("/api/auth/logout");
    await goto("/"); // TODO does not reload if we already in "/"
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
        <li class={segment === link.segment ? 'nav-item active' : 'nav-item'}>
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
    {segment}
    <span class="navbar-text">
      <a class="nav-link" href="." on:click={logout}>Log out</a>
    </span>
  </div>
</nav>
