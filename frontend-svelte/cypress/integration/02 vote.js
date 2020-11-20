describe('A validated user can vote', () => {
    before(() => {
        cy.exec("npm run db:vote");
    });

    beforeEach(() => {
        cy.visit('/', { onBeforeLoad: () => { localStorage.setItem('bella-ciao.lang', 'en') } });
        cy.get('#unique_id').type("88888888Y");
        cy.get('#password').type("12345678");
        cy.get('button').contains('Log in').click();
    });

    it('shows validated message and allows voting', () => {
        cy.get(".alert").contains('You have been validated');

        cy.get("#nonselected-candidates").find("tr").contains("candidate 1").siblings().find("button").click();
        cy.get("#nonselected-candidates").find("tr").contains("candidate 2").siblings().find("button").click();
        cy.get("#nonselected-candidates").find("tr").contains("candidate 3").siblings().find("button").click();

        cy.get("#selected-candidates").find("tr").contains("candidate 2").siblings().find("button").first().click();
        cy.get("#selected-candidates").find("tr").contains("candidate 3").siblings().find("button").last().click();

        cy.get("#candidates button").contains("Vote").click();
    });
});