Take a look at `new-metric-details.md` for details on the metrics to implement.

The end result will be 4 new charts in `$exerciseId.tsx`. 

They will be all barcharts will all four of these metrics:
- Total volume we already have. 
- Sessions best 1rm.
- Session average 1rm.
- Session average intensity.
- Session best intensity.

One or possible two tiles for `historial_1rm` and `historial_intensity`. 

Possibly two tiles for `session_best_1rm` and `session_best_intensity`.

Your job is to decide where to implement these new metrics in the `server` directory. 

My opinion is add a `historical_1rm` column to the `exercise` table. This will be updated at cycle milestones (≈8–12 weeks) or when a clearly higher reliable e1RM/tested 1RM is achieved; avoid frequent downgrades due to short-term variance. We could also implement `historial_intensity` if there is not too much overhead.

Questions I am pondering?
- Should we add a new column `1rm` to the `set` table? 
- Should we add a new column `intensity` to the `set` table although this depends on the newly added `1rm` column? 
- I imagine if we do add these new columns, we can compute on client side.
- Should we display these new metrics client side. 
- There are ui implications. I am hesitant to clutter the ui. We could have a settings to turn on/off these metrics.
- Not sure if we need to store a exercise's `average_1rm` and `average_intensity` for that corresponding day.
- Should they just be calculated on the fly?

These are just preliminary thoughts. I am open to suggestions. Do you have any questions. Take your time to think before you respond.