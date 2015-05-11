## Tasks

This folder contains all of the Gulp.js build tasks for both the UI. This includes bundling the client app and dependencies, building any styles, linting the code, etc.

The tasks are required in, in [gulpfile.js](../gulpfile.js).

Each file in the directory is a task you can run with `gulp {filename}`. This allows us to quickly see what tasks are currently available. Tasks should have names that describe the work they will do. The task and the file it is in should always have the same name.

For more on Gulp.js, visit [http://gulpjs.com/](http://gulpjs.com/)

Tasks (commonly used):

- dev: builds the development environment, watches files to run linting, run tests, and build styles or app bundle when files change
- test: runs the server and client tests one time
- tdd: watches js files for changes and runs tests when change event fires
- build-dev: builds the environment and files necessary to develop the application
- build-prod: builds the environment needed for production use
