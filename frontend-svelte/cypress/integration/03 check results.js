describe('An authenticated user can see the election results', () => {
    before(() => {
        cy.exec("npm run db:results");
    });

    beforeEach(() => {
        cy.visit('/', { onBeforeLoad: () => { localStorage.setItem('bella-ciao.lang', 'en') } });
        cy.get('#unique_id').type("88888888Y");
        cy.get('#password').type("12345678");
        cy.get('button').contains('Log in').click();
    });

    it('shows the election results', () => {
        cy.get("p").first().contains("The election has already ended");

        cy.get("table tbody td").contains("candidate 3").siblings().first().contains("4");
        cy.get("table tbody td").contains("candidate 1").siblings().first().contains("3");
        cy.get("table tbody td").contains("candidate 2").siblings().first().contains("0");
        cy.get("table tbody td").contains("candidate 4").siblings().first().contains("0");
    });
});