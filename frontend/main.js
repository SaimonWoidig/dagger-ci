if (typeof APP_CONFIG == "undefined") {
  document.addEventListener("configLoaded", main);
} else {
  main();
}

function main() {
  const apiBaseUrl = `${APP_CONFIG.api.proto}://${APP_CONFIG.api.host}:${APP_CONFIG.api.port}`;

  function refreshTask() {
    let taskDiv = document.getElementById("task");
    fetch(`${apiBaseUrl}/api/task`, { method: "GET" })
      .then((res) => res.json())
      .then(
        (resJson) =>
          (taskDiv.innerHTML = `${resJson.title}: ${resJson.completed}`)
      );
  }

  function updateTask(form) {
    let taskTitle = form.taskTitle.value;
    let taskCompleted = form.taskCompleted.checked;
    fetch(`${apiBaseUrl}/api/task`, {
      method: "PUT",
      body: JSON.stringify({ title: taskTitle, completed: taskCompleted }),
      headers: {
        "Content-Type": "application/json",
      },
    })
      .then((res) => {
        if (!res.ok) {
          throw res.statusText;
        }
        refreshTask();
      })
      .catch((err) => alert(`Error submitting: ${err}`))
      .finally(() => {
        form.taskTitle.value = "";
        form.taskCompleted.checked = false;
      });
  }

  document.addEventListener("DOMContentLoaded", refreshTask);

  let taskForm = document.getElementById("taskForm");
  taskForm.addEventListener("submit", (e) => {
    e.preventDefault();
    updateTask(taskForm);
  });
}
