document.addEventListener('DOMContentLoaded', function() {
    const subscribeButton = document.querySelector('.subscribe-button');
    const modal = document.getElementById('subscribe-modal');
    const closeButton = document.querySelector('.close-button');
    const subscribeForm = document.getElementById('subscribe-form');
    const modalServicesList = document.getElementById('modal-services-list');
    const messageElement = document.getElementById('subscription-message');

    // Show the modal
    subscribeButton.addEventListener('click', function(e) {
        e.preventDefault();
        // Fetch services and populate the modal before showing it
        fetch('/api/v1/status')
            .then(response => response.json())
            .then(data => {
                modalServicesList.innerHTML = ''; // Clear previous list
                if (data.services && data.services.length > 0) {
                    data.services.forEach(service => {
                        const checkboxDiv = document.createElement('div');
                        checkboxDiv.innerHTML = `
                            <input type="checkbox" id="service-${service.ID}" name="services" value="${service.ID}">
                            <label for="service-${service.ID}">${service.name}</label>
                        `;
                        modalServicesList.appendChild(checkboxDiv);
                    });
                } else {
                    modalServicesList.innerHTML = '<p>No services available for subscription.</p>';
                }
                modal.style.display = 'block';
            })
            .catch(error => {
                console.error('Error fetching services for modal:', error);
                alert('Could not load services. Please try again later.');
            });
    });

    // Hide the modal
    closeButton.addEventListener('click', function() {
        modal.style.display = 'none';
    });

    window.addEventListener('click', function(event) {
        if (event.target == modal) {
            modal.style.display = 'none';
        }
    });

    // Handle subscription form submission
    subscribeForm.addEventListener('submit', (event) => {
        event.preventDefault();
        const email = document.getElementById('email').value;
        const selectedServices = Array.from(document.querySelectorAll('input[name="services"]:checked')).map(cb => cb.value);
        const submitButton = subscribeForm.querySelector('button[type="submit"]');

        submitButton.disabled = true;
        messageElement.textContent = '';

        fetch('/api/v1/subscribers', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email: email, services: selectedServices.map(Number) }),
        })
        .then(async response => {
            if (response.ok) {
                messageElement.textContent = 'Successfully subscribed!';
                messageElement.style.color = 'green';
                setTimeout(() => {
                    modal.style.display = 'none';
                    messageElement.textContent = '';
                }, 2000);
            } else {
                const errorData = await response.json();
                messageElement.textContent = `Subscription failed: ${errorData.error}`;
                messageElement.style.color = 'red';
            }
        })
        .catch(error => {
            console.error('Error:', error);
            messageElement.textContent = 'An error occurred. Please try again.';
            messageElement.style.color = 'red';
        })
        .finally(() => {
            submitButton.disabled = false;
        });
    });

    const componentsList = document.getElementById('components-list');
    const incidentsList = document.getElementById('incidents-list');
    const maintenanceList = document.getElementById('maintenance-list');

    // Fetch status data from the API
    fetch('/api/v1/status')
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            // Clear any placeholder content
            componentsList.innerHTML = '';

            // Check if there are services to display
            if (data.services && data.services.length > 0) {
                data.services.forEach(service => {
                    const componentDiv = document.createElement('div');
                    componentDiv.className = 'component';

                    const componentName = document.createElement('span');
                    componentName.className = 'component-name';
                    componentName.textContent = service.name;

                    const componentStatus = document.createElement('span');
                    componentStatus.className = 'component-status operational'; // Assuming 'operational' for now
                    componentStatus.textContent = 'Operational';

                    componentDiv.appendChild(componentName);
                    componentDiv.appendChild(componentStatus);
                    componentsList.appendChild(componentDiv);
                });
            } else {
                componentsList.innerHTML = '<p>No services to display.</p>';
            }
        })
        .catch(error => {
            console.error('Error fetching status:', error);
            componentsList.innerHTML = '<p>Could not load status information.</p>';
        });

    // Fetch incidents data from the API
    fetch('/api/v1/incidents')
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            incidentsList.innerHTML = ''; // Clear placeholder

            if (data && data.length > 0) {
                data.forEach(incident => {
                    const incidentDiv = document.createElement('div');
                    incidentDiv.className = 'incident';

                    const incidentDate = new Date(incident.CreatedAt).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });

                    incidentDiv.innerHTML = `
                        <div class="incident-header">
                            <span class="incident-status ${incident.Status.toLowerCase()}">${incident.Status}</span>
                            <span class="incident-date">${incidentDate}</span>
                        </div>
                        <div class="incident-body">
                            <h4>${incident.Title}</h4>
                            <p>${incident.Description}</p>
                        </div>
                    `;
                    incidentsList.appendChild(incidentDiv);
                });
            } else {
                incidentsList.innerHTML = '<p>No incidents to display.</p>';
            }
        })
        .catch(error => {
            console.error('Error fetching incidents:', error);
            incidentsList.innerHTML = '<p>Could not load incident information.</p>';
        });

    // Fetch maintenance data from the API
    fetch('/api/v1/maintenance')
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            maintenanceList.innerHTML = ''; // Clear placeholder

            if (data && data.length > 0) {
                data.forEach(event => {
                    const eventDiv = document.createElement('div');
                    eventDiv.className = 'maintenance-event';

                    const startAt = new Date(event.StartAt).toLocaleString();
                    const endAt = new Date(event.EndAt).toLocaleString();

                    eventDiv.innerHTML = `
                        <h4>${event.Title}</h4>
                        <p>${event.Description}</p>
                        <p><strong>Status:</strong> ${event.Status}</p>
                        <p><strong>Scheduled for:</strong> ${startAt} to ${endAt}</p>
                    `;
                    maintenanceList.appendChild(eventDiv);
                });
            } else {
                maintenanceList.innerHTML = '<p>No scheduled maintenance.</p>';
            }
        })
        .catch(error => {
            console.error('Error fetching maintenance events:', error);
            maintenanceList.innerHTML = '<p>Could not load maintenance information.</p>';
        });

    // Admin Dashboard - Monitor Management
    const monitorsTable = document.getElementById('monitors-table');
    if (monitorsTable) {
        const monitorModal = document.getElementById('monitor-modal');
        const addMonitorBtn = document.getElementById('add-monitor-btn');
        const monitorForm = document.getElementById('monitor-form');
        const monitorModalClose = monitorModal.querySelector('.close-button');

        const openMonitorModal = (monitor = null) => {
            monitorForm.reset();
            if (monitor) {
                document.getElementById('monitor-id').value = monitor.ID;
                document.getElementById('monitor-name').value = monitor.name;
                document.getElementById('monitor-url').value = monitor.url;
                document.getElementById('monitor-type').value = monitor.type;
                document.getElementById('monitor-interval').value = monitor.interval;
                document.getElementById('monitor-expected-status').value = monitor.expected_status;
            } else {
                document.getElementById('monitor-id').value = '';
            }
            monitorModal.style.display = 'block';
        };

        const closeMonitorModal = () => {
            monitorModal.style.display = 'none';
        };

        const loadMonitors = async () => {
            try {
                const response = await fetch('/api/v1/admin/monitors', {
                    credentials: 'include'
                });
                if (!response.ok) throw new Error('Failed to fetch monitors');
                const monitors = await response.json();

                const tbody = monitorsTable.querySelector('tbody');
                tbody.innerHTML = '';
                if (monitors && monitors.length > 0) {
                    monitors.forEach(monitor => {
                        const row = document.createElement('tr');
                        row.innerHTML = `
                            <td>${monitor.name}</td>
                            <td>${monitor.url}</td>
                            <td>${monitor.interval}</td>
                            <td class="actions">
                                <button class="edit-monitor" data-id="${monitor.ID}">Edit</button>
                                <button class="delete-monitor" data-id="${monitor.ID}">Delete</button>
                            </td>
                        `;
                        tbody.appendChild(row);
                    });
                } else {
                    tbody.innerHTML = '<tr><td colspan="4">No monitors configured.</td></tr>';
                }
            } catch (error) {
                console.error('Error loading monitors:', error);
            }
        };

        monitorForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const id = document.getElementById('monitor-id').value;
            const method = id ? 'PUT' : 'POST';
            const url = id ? `/api/v1/admin/monitors/${id}` : '/api/v1/admin/monitors';

            const formData = new FormData(monitorForm);
            const data = {
                name: formData.get('name'),
                url: formData.get('url'),
                type: formData.get('type'),
                interval: parseInt(formData.get('interval'), 10),
                expected_status: parseInt(formData.get('expected_status'), 10),
            };

            try {
                const response = await fetch(url, {
                    method,
                    headers: { 'Content-Type': 'application/json' },
                    credentials: 'include',
                    body: JSON.stringify(data),
                });

                if (!response.ok) {
                    const error = await response.json();
                    throw new Error(error.details || 'Failed to save monitor');
                }

                closeMonitorModal();
                loadMonitors();
            } catch (error) {
                alert(`Error: ${error.message}`);
            }
        });

        monitorsTable.addEventListener('click', async (e) => {
            if (e.target.classList.contains('edit-monitor')) {
                const id = e.target.dataset.id;
                const response = await fetch(`/api/v1/admin/monitors`, {
                    credentials: 'include'
                }); // Fetch all and find
                const monitors = await response.json();
                const monitor = monitors.find(m => m.ID == id);
                if (monitor) openMonitorModal(monitor);
            }

            if (e.target.classList.contains('delete-monitor')) {
                const id = e.target.dataset.id;
                if (confirm('Are you sure you want to delete this monitor?')) {
                    try {
                        const response = await fetch(`/api/v1/admin/monitors/${id}`, { 
                            method: 'DELETE',
                            credentials: 'include'
                        });
                        if (!response.ok) throw new Error('Failed to delete monitor');
                        loadMonitors();
                    } catch (error) {
                        alert(`Error: ${error.message}`);
                    }
                }
            }
        });

        addMonitorBtn.addEventListener('click', () => openMonitorModal());
        monitorModalClose.addEventListener('click', closeMonitorModal);

        // Initial load
        loadMonitors();
    }
});
