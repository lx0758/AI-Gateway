## MODIFIED Requirements

### Requirement: Allow manual model deletion

The system SHALL allow deleting manually created models AND reject deletion of synced models.

#### Scenario: Delete manual model
- **WHEN** admin deletes a model with source="manual"
- **THEN** system removes the model from database

#### Scenario: Reject delete sync model
- **WHEN** admin attempts to delete a model with source="sync"
- **THEN** system returns an error with status 400
- **AND** system does NOT remove the model from database

#### Scenario: Error message for sync model deletion
- **WHEN** admin attempts to delete a model with source="sync"
- **THEN** system returns error message "cannot delete synced model, only manually added models can be deleted"
