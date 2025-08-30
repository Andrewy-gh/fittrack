Can we create a new endpoint to retrieve all workout focus values for each user? This was a query I had in mind:
```sql
SELECT DISTINCT workout_focus 
FROM workouts 
WHERE user_id = ? 
  AND workout_focus IS NOT NULL
ORDER BY workout_focus;
```
This is a preliminary query, we can iterate depending on what the code base looks like. I think we create an index for workout focus and user, so if the query can be optimized with the index, please let me know.

Next steps:
- [ ] Add query to `query.sql`
- [ ] Create a new endpoint to retrieve all workout focus values for each user
- [ ] Update the swagger documentation
- [ ] Update or create the tests
- [ ] Update or create  the models

Think about your answer first before you respond. "Do you have any question before you implement this?"