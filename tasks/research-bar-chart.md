Do some research on potential TypeScript libraries that support bar charts. I want you to create a bar chart that replicates the look of `chart-bar-vol.tsx`. Here is what the current component looks like:
`IMG_8993.jpg` 

The issue I am having is the `Brush` component that selects the range for the data does not have a great UX. It is hard to control on mobile.

Keys to success:
- Ability to style (use themes in `styles.css`) so that it has the same style of the current `client` app.
- Ability to easily modify the range
- Offer a tooltip when a bar is pressed it shows the data already shown in `chart-bar-vol.tsx` **Not a priority but nice to have**

**Some possible design alternatives:**
This is a bar chart from apple's health app: `Steps.png`. It has button on the top which decide how the data is displayed. Currently `M` is highlighted for month. As the range increases the bars get smaller. It's a widely known app so I'm sure you can research the UI/UX of it. It also allows the user to horizontally scroll left and right. This might present some data fetching issues though.

**Steps**
- Do research on up to 3 alternatives. 5 max.
- Make a plan on how you would implement these components
- Have all these components shown in `App.tsx` of the `chart-test` directory.
- Keep the same styles, spacing as `bar-chart-vol.tsx` so I can compare.
- Save this plan in a markdown file as `tasks-bar-chart-plan.md`

Take your time and ultrathink before responding. If you have any unresolved questions, save for the end of any. Be hyper concise in your response. Sacrifice for grammar for concision.
