/// <reference types="Cypress" />

import { faker } from '@faker-js/faker';
import { getKey } from '../support/utils';

context('Delete', () => {
  beforeEach(() => {
    cy.visit('/');
  });

  it('should delete an existing url', () => {
    const key = getKey();
    cy.addUrl({ key, url: faker.internet.url() });
    cy.getResult(key).find('[data-e2e="delete"]').click();

    cy.getHandle('delete-modal').contains(`Delete ${key}?`);
    cy.getHandle('delete-modal').contains(
      `Are you sure you want to delete "${key}"? This cannot be undone.`,
    );
    cy.getHandle('delete-confirm').click();

    cy.getHandle('alert').contains(`Successfully deleted ${key}`);
    cy.getHandle('Search Results').should('not.contain', key);
    cy.request(`/api/url/${key}`).its('body').should('equal', null);
  });
});
