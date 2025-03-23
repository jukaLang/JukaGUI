const canvas = document.getElementById('canvas');
const output = document.getElementById('output');
const datax = document.getElementById('datax');
const datay = document.getElementById('datay');
const colorPicker = document.getElementById('colorPicker');
const bgColorPicker = document.getElementById('bgColorPicker');
const fontSizePicker = document.getElementById('fontSizePicker');
const triggerSelector = document.getElementById('triggerSelector');
const sceneChangeSelector = document.getElementById('sceneChangeSelector');
const externalAppPath = document.getElementById('externalAppPath');
const variableChangeSelector = document.getElementById('variableChangeSelector');
const variableChangeValue = document.getElementById('variableChangeValue');
const backgroundFileInput = document.getElementById('backgroundFile');
const canvasSizeSelect = document.getElementById('canvasSize');
const sceneSelector = document.getElementById('sceneSelector');
const titleSizeInput = document.getElementById('titleSize');
const bigSizeInput = document.getElementById('bigSize');
const mediumSizeInput = document.getElementById('mediumSize');
const smallSizeInput = document.getElementById('smallSize');
const variableSelector = document.getElementById('variableSelector');
const customWidthInput = document.getElementById('customWidth');
const customHeightInput = document.getElementById('customHeight');
let backgroundPath = '';
let scenes = { 'Scene 1': [] };
let currentScene = 'Scene 1';
let variables = {};


function toggleCustomFields() {
    const select = document.getElementById('canvasSize');
    const customFields = document.getElementById('customSizeFields');
    
    if (select.value === 'custom') {
        customFields.style.display = 'inline-block';
    } else {
        customFields.style.display = 'none';
    }
}

function updateCanvasSize() {
    const canvasSize = canvasSizeSelect.value;
    
    if (canvasSize === 'custom') {
        
        // Get width and height from custom input fields
        const width = parseInt(customWidthInput.value) || 1280; // Default if empty
        const height = parseInt(customHeightInput.value) || 720; // Default if empty
        
        canvas.style.width = `${width}px`;
        canvas.style.height = `${height}px`;
    } else {
        // Handle preset sizes
        const [width, height] = canvasSize.split('x').map(Number);
        canvas.style.width = `${width}px`;
        canvas.style.height = `${height}px`;
    }
    toggleCustomFields();
}

function updateScenes() {
    // Save current elements to the current scene
    if (!scenes[currentScene]) {
        scenes[currentScene] = [];
    }
    scenes[currentScene] = Array.from(canvas.children).map(el => el.cloneNode(true));
    console.log(`Scene ${currentScene} updated with elements:`, scenes[currentScene]);

    // Clear the canvas
    canvas.innerHTML = '';

    // Re-render elements for the current scene
    scenes[currentScene].forEach(el => {
        canvas.appendChild(el);
    });
    
    // Update font sizes if necessary
    updateElementFontSizes();
}



function changeScene() {
    updateScenes();
    currentScene = sceneSelector.value;
    loadScene(currentScene);
    updateElementFontSizes();
    updateVariableText();
    document.querySelectorAll('.menu').forEach(menuEl => {
        updateMenuSceneButtons(menuEl);
    });
    
}

function loadScene(sceneName) {
    canvas.innerHTML = '';
    if (scenes[sceneName]) {
        scenes[sceneName].forEach(el => {
            const clonedEl = el.cloneNode(true);
            setupElementEvents(clonedEl);
            canvas.appendChild(clonedEl);
        });
    } else {
        console.error(`Scene ${sceneName} not found`);
    }
}


function addScene() {
    const newSceneName = 'Scene ' + (Object.keys(scenes).length + 1);
    scenes[newSceneName] = [];
    const option = document.createElement('option');
    option.value = newSceneName;
    option.textContent = newSceneName;
    sceneSelector.appendChild(option);
    sceneSelector.value = newSceneName;
    updateSceneChangeSelector();
    changeScene();
    
    // Add menu to new scene
    const canvasHeight = canvas.offsetHeight;
    addElement('menu', 0, canvasHeight - 50);
    
    document.querySelectorAll('.menu').forEach(menuEl => {
        updateMenuSceneButtons(menuEl);
    });
}

function renameScene() {
    const newName = prompt('Enter new scene name:', currentScene);
    if (newName && !scenes[newName]) {
        scenes[newName] = scenes[currentScene];
        delete scenes[currentScene];
        currentScene = newName;

        // Update the option in the scene selector
        const selectedOption = sceneSelector.querySelector(`option[value="${sceneSelector.value}"]`);
        if (selectedOption) {
            selectedOption.value = newName;
            selectedOption.textContent = newName;
        }

        // Reflect the change in the scene selector
        sceneSelector.querySelector(`option[value="${sceneSelector.value}"]`).value = newName;
        sceneSelector.querySelector(`option[value="${sceneSelector.value}"]`).textContent = newName;

        sceneSelector.value = newName;
        updateSceneChangeSelector();
        loadScene(newName);
        console.log(`Scene renamed to: ${newName}`);
    } else {
        console.log('Scene name already exists or invalid input.');
    }
}

function deleteScene() {
    if (Object.keys(scenes).length <= 1) {
        alert('Cannot delete the only scene. You must have at least one scene.');
        return;
    }

    if (confirm(`Are you sure you want to delete scene "${currentScene}"?`)) {
        // Get the current index to determine which scene to show next
        const sceneNames = Object.keys(scenes);
        const currentIndex = sceneNames.indexOf(currentScene);
        
        // Remove the scene from scenes object
        delete scenes[currentScene];
        
        // Remove the option from scene selector
        const optionToRemove = sceneSelector.querySelector(`option[value="${currentScene}"]`);
        if (optionToRemove) {
            sceneSelector.removeChild(optionToRemove);
        }
        
        // Determine which scene to show next
        let nextSceneIndex = currentIndex > 0 ? currentIndex - 1 : 0;
        if (nextSceneIndex >= sceneNames.length - 1) {
            nextSceneIndex = sceneNames.length - 2; // Adjust for removal
        }
        
        const nextScene = sceneNames[nextSceneIndex] === currentScene ? 
            sceneNames[(nextSceneIndex + 1) % (sceneNames.length - 1)] : 
            sceneNames[nextSceneIndex];
            
        // Update current scene and load it
        currentScene = nextScene;
        sceneSelector.value = currentScene;
        
        // Update selectors and UI
        updateSceneChangeSelector();
        loadScene(currentScene);
        
        // Update menus
        document.querySelectorAll('.menu').forEach(menuEl => {
            updateMenuSceneButtons(menuEl);
        });
        
        console.log(`Scene "${currentScene}" deleted successfully.`);
    }
}

function duplicateScene() {
    const newSceneName = prompt('Enter name for the duplicated scene:', `${currentScene} Copy`);
    
    if (newSceneName && !scenes[newSceneName]) {
        // Clone current scene elements
        scenes[newSceneName] = scenes[currentScene].map(el => el.cloneNode(true));
        
        // Add new scene to selector
        const option = document.createElement('option');
        option.value = newSceneName;
        option.textContent = newSceneName;
        sceneSelector.appendChild(option);
        
        // Switch to the new scene
        sceneSelector.value = newSceneName;
        currentScene = newSceneName;
        
        // Update UI and selectors
        updateSceneChangeSelector();
        loadScene(newSceneName);
        
        // Update menus
        document.querySelectorAll('.menu').forEach(menuEl => {
            updateMenuSceneButtons(menuEl);
        });
        
        console.log(`Scene duplicated as "${newSceneName}"`);
    } else if (scenes[newSceneName]) {
        alert('A scene with that name already exists. Please choose a different name.');
    }
}


function updateSceneChangeSelector() {
    sceneChangeSelector.innerHTML = '';
    Object.keys(scenes).forEach(sceneName => {
        const option = document.createElement('option');
        option.value = sceneName;
        option.textContent = sceneName;
        sceneChangeSelector.appendChild(option);
    });
}

document.addEventListener('DOMContentLoaded', () => {
    setupToolbar();
    setupFontSizeControls();
    backgroundFileInput.addEventListener('change', setBackground);
    canvasSizeSelect.addEventListener('change', updateCanvasSize);
    setupVariableControls();
    updateSceneChangeSelector();
    updateCanvasSize();
    
    // Add "Scene 1" to the scene selector
    const option = document.createElement('option');
    option.value = 'Scene 1';
    option.textContent = 'Scene 1';
    sceneSelector.appendChild(option);
    sceneSelector.value = 'Scene 1';

    // Add menu to initial scene
    setTimeout(() => { // Wait for canvas size to update
        const canvasHeight = canvas.offsetHeight;
        addElement('menu', 0, canvasHeight - 50);
    }, 0);
});

function setupToolbar() {
    document.querySelectorAll('.toolbar .element').forEach(el => {
        el.addEventListener('dragstart', event => {
            event.dataTransfer.setData('type', event.target.getAttribute('data-type'));
        });
    });
}

function setupFontSizeControls() {
    const fontSizeInputs = document.querySelectorAll('.font-size-controls input');
    fontSizeInputs.forEach(input => {
        input.addEventListener('input', () => {
            updateElementFontSizes();
            updateScenes();
        });
    });
}

function setupVariableControls() {
    variableSelector.addEventListener('change', changeVariable);
    variableChangeSelector.addEventListener('change', () => {
        variableChangeValue.style.display = 'block';
    });
}

function addVariable() {
    const variableName = prompt('Enter variable name:');
    if (variableName && !variables[variableName]) {
        variables[variableName] = '';
        const option = document.createElement('option');
        option.value = variableName;
        option.textContent = variableName;
        variableSelector.appendChild(option);
        const variableOption = document.createElement('option');
        variableOption.value = variableName;
        variableOption.textContent = variableName;
        variableChangeSelector.appendChild(variableOption);
        variableSelector.value = variableName;
        showVariableControls();
        changeVariable();
    }
}

function renameVariable() {
    const newName = prompt('Enter new variable name:', variableSelector.value);
    if (newName && !variables[newName]) {
        variables[newName] = variables[variableSelector.value];
        delete variables[variableSelector.value];

        const selectedOption = variableSelector.querySelector(`option[value="${variableSelector.value}"]`);
        selectedOption.value = newName;
        selectedOption.textContent = newName;

        const selectedChangeOption = variableChangeSelector.querySelector(`option[value="${variableSelector.value}"]`);
        selectedChangeOption.value = newName;
        selectedChangeOption.textContent = newName;

        variableSelector.value = newName;
        variableChangeSelector.value = newName;
        changeVariable();
    }
}

// Function to change variable
function changeVariable() {
    const variableName = variableSelector.value;
    document.getElementById('variableValueInput').value = variables[variableName];
}

// Function to show variable controls
function showVariableControls() {
    document.getElementById('variableSelector').style.display = 'inline';
    document.getElementById('variableValueInput').style.display = 'inline';
    document.getElementById('renameVariableButton').style.display = 'inline';
    document.getElementById('changeValueButton').style.display = 'inline';
}

// Function to show Change Value button and input field
function showChangeValue() {
    const variableName = variableSelector.value;
    const variableValue = prompt('Enter new variable value:', variables[variableName]);
    if (variableValue !== null) {
        variables[variableName] = variableValue;
        document.getElementById('variableValueInput').value = variables[variableName];
        updateVariableText();
    }
}

function showTooltip(event) {
    const target = event.target;
    const variableNames = target.getAttribute('data-variable');
    if (variableNames) {
        let tooltip = document.createElement('div');
        tooltip.className = 'tooltip';
        let tooltipText = '';
        const addedVariables = new Set(); // To track added variables and avoid duplicates
        variableNames.split(',').forEach(variableName => {
            if (variables[variableName] !== undefined && !addedVariables.has(variableName)) {
                tooltipText += `$${variableName}:${variables[variableName]} `;
                addedVariables.add(variableName);
            }
        });
        tooltip.textContent = tooltipText.trim();
        document.body.appendChild(tooltip);
        const rect = target.getBoundingClientRect();
        tooltip.style.left = `${rect.left + window.scrollX}px`;
        tooltip.style.top = `${rect.top + window.scrollY - tooltip.offsetHeight}px`;
        target._tooltip = tooltip;
    }
}


function hideTooltip(event) {
    const target = event.target;
    if (target._tooltip) {
        document.body.removeChild(target._tooltip);
        delete target._tooltip;
    }
}

function setupHoverEvents(el) {
    el.addEventListener('mouseenter', showTooltip);
    el.addEventListener('mouseleave', hideTooltip);
}

// Event listener for variable selector
variableSelector.addEventListener('change', changeVariable);

function updateElementFontSizes() {
    const elements = document.querySelectorAll('.canvas .element');
    elements.forEach(el => {
        const fontType = el.getAttribute('data-font');
        el.style.fontSize = `${getFontSize(fontType)}px`;
    });
}

function updateVariableText() {
    const elements = document.querySelectorAll('.canvas .element');
    elements.forEach(el => {
        if (el.getAttribute('data-type') === 'button' || el.getAttribute('data-type') === 'label' || el.getAttribute('data-type') === 'input' || el.getAttribute('data-type') === 'image') {
            let textSpan = el.querySelector('.text-content');
            let text = textSpan.textContent;
            // Keep $variable names and show tooltip on hover
            const addedVariables = new Set(); // To track added variables and avoid duplicates
            Object.keys(variables).forEach(key => {
                const variablePlaceholder = `$${key}`;
                if (text.includes(variablePlaceholder) && !addedVariables.has(key)) {
                    addedVariables.add(key);
                    el.setAttribute('data-variable', `${el.getAttribute('data-variable') ? el.getAttribute('data-variable') + ',' : ''}${key}`);
                    setupHoverEvents(el);
                }
            });
            textSpan.textContent = text; // Ensure text remains unchanged
        }
    });
}


canvas.addEventListener('dragover', event => {
    event.preventDefault();
});

canvas.addEventListener('drop', event => {
    event.preventDefault();
    const type = event.dataTransfer.getData('type');
    const canvasRect = canvas.getBoundingClientRect();
    const x = event.clientX - canvasRect.left;
    const y = event.clientY - canvasRect.top;
    addElement(type, x, y);
});

function setupElementEvents(el) {
    setupHoverEvents(el);

    el.addEventListener('mousedown', event => {
        if (event.button === 2) { // Right mouse button
            el.style.cursor = 'nwse-resize';
            const startX = event.clientX;
            const startY = event.clientY;
            const startWidth = el.offsetWidth;
            const startHeight = el.offsetHeight;

            function onMouseMove(event) {
                const newWidth = startWidth + (event.clientX - startX);
                const newHeight = startHeight + (event.clientY - startY);
                el.style.width = `${newWidth}px`;
                el.style.height = `${newHeight}px`;
                el.setAttribute('data-width', newWidth);
                el.setAttribute('data-height', newHeight);
            }

            function onMouseUp() {
                el.style.cursor = 'grab';
                document.removeEventListener('mousemove', onMouseMove);
                document.removeEventListener('mouseup', onMouseUp);
            }

            document.addEventListener('mousemove', onMouseMove);
            document.addEventListener('mouseup', onMouseUp);
        } else { // Left mouse button
            el.style.cursor = 'grabbing';
            const xOffset = event.clientX - el.offsetLeft;
            const yOffset = event.clientY - el.offsetTop;

            function onMouseMove(event) {
                let newX = event.clientX - xOffset;
                let newY = event.clientY - yOffset;

                // Prevent dragging outside the canvas
                const canvasRect = canvas.getBoundingClientRect();
                const elRect = el.getBoundingClientRect();

                if (newX < 0) newX = 0;
                if (newY < 0) newY = 0;
                if (newX + elRect.width > canvasRect.width) newX = canvasRect.width - elRect.width;
                if (newY + elRect.height > canvasRect.height) newY = canvasRect.height - elRect.height;

                newX = parseInt(newX)
                newY = parseInt(newY)

                el.style.left = `${newX}px`;
                el.style.top = `${newY}px`;
                el.setAttribute('data-x', newX);
                el.setAttribute('data-y', newY);
            }

            function onMouseUp() {
                el.style.cursor = 'grab';
                document.removeEventListener('mousemove', onMouseMove);
                document.removeEventListener('mouseup', onMouseUp);
            }

            document.addEventListener('mousemove', onMouseMove);
            document.addEventListener('mouseup', onMouseUp);
        }
    });

    el.addEventListener('contextmenu', event => {
        event.preventDefault(); // Prevent the context menu from appearing
    });

    el.addEventListener('dblclick', event => {
        event.stopPropagation(); // Prevent the click from bubbling up
        const textSpan = el.querySelector('.text-content');
        if (el.getAttribute('data-type') === 'button' || el.getAttribute('data-type') === 'label' || el.getAttribute('data-type') === 'input') {
            const newText = prompt("Enter new text:", textSpan.textContent);
            if (newText !== null) {
                textSpan.textContent = newText; // Change only the text span content
                updateVariableText();
            }
        } else if (el.getAttribute('data-type') === 'image') {
            const fileInput = document.createElement('input');
            fileInput.type = 'file';
            fileInput.accept = 'image/*';
            fileInput.style.display = 'none';
            document.body.appendChild(fileInput);

            fileInput.onchange = function(event) {
                const file = event.target.files[0];
                if (file) {
                    const reader = new FileReader();
                    reader.onload = function(event) {
                        el.style.backgroundImage = `url(${event.target.result})`;
                        el.style.backgroundSize = 'cover';
                        el.setAttribute('data-image', file.name);
                    };
                    reader.readAsDataURL(file);
                }
            };

            fileInput.click();
            document.body.removeChild(fileInput);
        }
    });

    el.addEventListener('click', () => {
        hideControls(); // Hide all controls first

        document.querySelector('.control-label[for="datax"]').style.display = 'block';
        datax.style.display = 'block';
        datax.innerText = el.getAttribute('data-x');

        document.querySelector('.control-label[for="datay"]').style.display = 'block';
        datay.style.display = 'block';
        datay.innerText = el.getAttribute('data-y');


        if (el.getAttribute('data-type') !== 'image' && el.getAttribute('data-type') !== "input") {
            document.querySelector('.control-label[for="colorPicker"]').style.display = 'block';
            colorPicker.value = el.getAttribute('data-color') || '#000000';
            colorPicker.style.display = 'block';
            colorPicker.oninput = function() {
                el.style.color = colorPicker.value;
                el.setAttribute('data-color', colorPicker.value);
            };

            document.querySelector('.control-label[for="fontSizePicker"]').style.display = 'block';
            fontSizePicker.value = el.getAttribute('data-font') || 'medium';
            fontSizePicker.style.display = 'block';
            fontSizePicker.onchange = function() {
                el.setAttribute('data-font', fontSizePicker.value);
                el.style.fontSize = getFontSize(fontSizePicker.value) + 'px';
            };

            if (el.getAttribute('data-type') !== 'label') {
            document.querySelector('.control-label[for="bgColorPicker"]').style.display = 'block';
            bgColorPicker.value = el.getAttribute('data-bg-color') || '#ffffff';
            bgColorPicker.style.display = 'block';
            bgColorPicker.oninput = function() {
                el.style.backgroundColor = bgColorPicker.value;
                el.setAttribute('data-bg-color', bgColorPicker.value);
            };
        }

        if (el.getAttribute('data-type') !== 'label') {
            document.querySelector('.control-label[for="triggerSelector"]').style.display = 'block';
            triggerSelector.value = el.getAttribute('data-trigger') || '';
            triggerSelector.style.display = 'block';
            triggerSelector.onchange = function() {
                el.setAttribute('data-trigger', triggerSelector.value);
                // Hide all trigger-related fields initially
                sceneChangeSelector.style.display = 'none';
                externalAppPath.style.display = 'none';
                variableChangeSelector.style.display = 'none';
                variableChangeValue.style.display = 'none';
                
                if (triggerSelector.value === 'change_scene') {
                    sceneChangeSelector.style.display = 'block';
                    sceneChangeSelector.value = el.getAttribute('data-trigger-target') || '';
                    sceneChangeSelector.onchange = function() {
                        el.setAttribute('data-trigger-target', sceneChangeSelector.value);
                    };
                } else if (triggerSelector.value === 'external_app') {
                    externalAppPath.style.display = 'block';
                    externalAppPath.value = el.getAttribute('data-trigger-target') || '';
                    externalAppPath.oninput = function() {
                        el.setAttribute('data-trigger-target', externalAppPath.value);
                    };
                    variableChangeValue.style.display = 'block';
                    variableChangeValue.value = el.getAttribute('data-trigger-value') || '';
                    variableChangeValue.oninput = function() {
                        el.setAttribute('data-trigger-value', variableChangeValue.value);
                    };
                } else if (triggerSelector.value === 'set_variable') {
                    variableChangeSelector.style.display = 'block';
                    variableChangeValue.style.display = 'block';
                    variableChangeSelector.value = el.getAttribute('data-trigger-target') || '';
                    variableChangeValue.value = el.getAttribute('data-trigger-value') || '';
                } else if (triggerSelector.value === 'play_video') {
                    externalAppPath.style.display = 'block';
                    externalAppPath.value = el.getAttribute('data-trigger-target') || '';
                } else if (triggerSelector.value === 'play_image') {
                    externalAppPath.style.display = 'block';
                    externalAppPath.value = el.getAttribute('data-trigger-target') || '';
                }
            };

            // Set initial state based on the current trigger
            if (el.getAttribute('data-trigger') === 'change_scene') {
                sceneChangeSelector.style.display = 'block';
                sceneChangeSelector.value = el.getAttribute('data-trigger-target') || '';
            } else if (el.getAttribute('data-trigger') === 'external_app') {
                externalAppPath.style.display = 'block';
                externalAppPath.value = el.getAttribute('data-trigger-target') || '';
            } else if (el.getAttribute('data-trigger') === 'set_variable') {
                variableChangeSelector.style.display = 'block';
                variableChangeValue.style.display = 'block';
                variableChangeSelector.value = el.getAttribute('data-trigger-target') || '';
                variableChangeValue.value = el.getAttribute('data-trigger-value') || '';
            }else if (el.getAttribute('data-trigger') === 'play_video') {
                externalAppPath.style.display = 'block';
                externalAppPath.value = el.getAttribute('data-trigger-target') || '';
            }
            else if (el.getAttribute('data-trigger') === 'play_image') {
                externalAppPath.style.display = 'block';
                externalAppPath.value = el.getAttribute('data-trigger-target') || '';
            }
        }
        } else {
            document.querySelector('.control-label[for="colorPicker"]').style.display = 'none';
            colorPicker.style.display = 'none';
            document.querySelector('.control-label[for="fontSizePicker"]').style.display = 'none';
            fontSizePicker.style.display = 'none';
            document.querySelector('.control-label[for="bgColorPicker"]').style.display = 'none';
            bgColorPicker.style.display = 'none';
            document.querySelector('.control-label[for="triggerSelector"]').style.display = 'none';
            triggerSelector.style.display = 'none';
        }
    });

    el.addEventListener('contextmenu', event => {
        event.preventDefault(); // Prevent the context menu from appearing
    });
    
    const removeButton = el.querySelector('.remove-button');
    if (removeButton) {
        removeButton.addEventListener('click', (event) => {
            event.stopPropagation();
            const parent = el.parentNode;
            parent.removeChild(el);
            const sceneElements = scenes[currentScene];
            const elementIndex = sceneElements.indexOf(el);
            if (elementIndex > -1) {
                sceneElements.splice(elementIndex, 1);
            }
        });
    }
}



function addElement(type, x, y) {
  const el = document.createElement('div');
  el.classList.add('element');
  el.style.position = 'absolute';
  el.style.left = `${x}px`;
  el.style.top = `${y}px`;

  // Setting default width and height
  let defaultWidth = 'auto';
  let defaultHeight = type === 'menu' ? '50px' : 'auto';
  
  if (type === 'image') {
    defaultWidth = '100px';
    defaultHeight = '100px';
  } else if (type === 'menu') {
    defaultWidth = '100%';
    // Position at bottom of canvas
    const canvasRect = canvas.getBoundingClientRect();
    y = canvasRect.height - parseInt(defaultHeight);
    el.style.top = `${y}px`;
    el.style.left = '0px';
  }

  // Set width and height
  el.style.width = defaultWidth;
  el.style.height = defaultHeight;

  // Add content based on element type
  if (type === 'menu') {
    // Create menu structure
    el.innerHTML = `
      <div class="handle"></div>
      <div class="menu-scene-buttons"></div>
      <button class="menu-language">EN</button>
      <div class="menu-clock">00:00</div>
    `;
    el.classList.add('menu');
    
    // Initialize the menu
    initializeMenu(el);
    
    // Start the clock
    updateMenuClock(el.querySelector('.menu-clock'));
  } else {
    // For non-menu elements, add the text content and remove button as before
    const textSpan = document.createElement('span');
    textSpan.classList.add('text-content');
    textSpan.textContent = type.charAt(0).toUpperCase() + type.slice(1);
    el.appendChild(textSpan);
    
    const removeButton = document.createElement('span');
    removeButton.textContent = '✕';
    removeButton.classList.add('remove-button');
    el.appendChild(removeButton);
  }

  el.setAttribute('data-type', type);
  el.setAttribute('data-x', x | 0);
  el.setAttribute('data-y', y | 0);
  
  if (type !== 'menu') {
    el.setAttribute('data-color', '#000000'); // Default text color
    el.setAttribute('data-bg-color', type === 'button' ? '#ffffff' : ''); // Default background color for buttons
    el.setAttribute('data-font', 'medium'); // Default font size
    
    if (type === 'image') {
      el.setAttribute('data-width', defaultWidth);
      el.setAttribute('data-height', defaultHeight);
    }
    
    el.style.fontSize = '24px'; // Setting a default font size for buttons and labels
    el.style.padding = '5px'; // Adjusting padding to make elements fit the text size
    
    if (type === 'label') {
      el.style.background = 'none'; // No background for labels
    }
  } else {
    el.setAttribute('data-height', defaultHeight.replace('px', ''));
  }

  // Attach event listeners
  if (type === 'menu') {
    setupMenuEvents(el);
  } else {
    setupElementEvents(el);
  }

  // Append element to canvas
  canvas.appendChild(el);

  // Save to the current scene
  if (!scenes[currentScene]) {
    scenes[currentScene] = [];
  }
  scenes[currentScene].push(el.cloneNode(true));

  return el;
}





function initializeMenu(menuEl) {
  updateMenuSceneButtons(menuEl);
}

// Update the scene buttons in the menu
function updateMenuSceneButtons(menuEl) {
  const sceneButtonsContainer = menuEl.querySelector('.menu-scene-buttons');
  sceneButtonsContainer.innerHTML = '';
  
  Object.keys(scenes).forEach(sceneName => {
    const button = document.createElement('button');
    button.classList.add('menu-scene-button');
    if (sceneName === currentScene) {
      button.classList.add('active');
    }
    button.textContent = sceneName;
    button.addEventListener('click', () => {
      sceneSelector.value = sceneName;
      changeScene();
      
      // Update active button
      menuEl.querySelectorAll('.menu-scene-button').forEach(btn => {
        btn.classList.remove('active');
      });
      button.classList.add('active');
    });
    sceneButtonsContainer.appendChild(button);
  });
}

// Update the clock in the menu
function updateMenuClock(clockEl) {
  const updateTime = () => {
    const now = new Date();
    const hours = now.getHours().toString().padStart(2, '0');
    const minutes = now.getMinutes().toString().padStart(2, '0');
    clockEl.textContent = `${hours}:${minutes}`;
  };
  
  updateTime();
  setInterval(updateTime, 6000); // Update every minute
}


// Set up event listeners for the menu
function setupMenuEvents(menuEl) {
  const handle = menuEl.querySelector('.handle');
  
  handle.addEventListener('mousedown', event => {
    event.stopPropagation();
    menuEl.style.cursor = 'grabbing';
    const startY = event.clientY;
    const startTop = parseInt(menuEl.style.top);
    
    function onMouseMove(event) {
      const canvasRect = canvas.getBoundingClientRect();
      const newTop = Math.max(0, Math.min(canvasRect.height - menuEl.offsetHeight, startTop + (event.clientY - startY)));
      menuEl.style.top = `${newTop}px`;
      menuEl.setAttribute('data-y', newTop);
    }
    
    function onMouseUp() {
      menuEl.style.cursor = '';
      document.removeEventListener('mousemove', onMouseMove);
      document.removeEventListener('mouseup', onMouseUp);
    }
    
    document.addEventListener('mousemove', onMouseMove);
    document.addEventListener('mouseup', onMouseUp);
  });
  
  // Add the remove button
  const removeButton = document.createElement('span');
  removeButton.textContent = '✕';
  removeButton.classList.add('remove-button');
  menuEl.appendChild(removeButton);
  
  removeButton.addEventListener('click', (event) => {
    event.stopPropagation();
    const parent = menuEl.parentNode;
    parent.removeChild(menuEl);
    const sceneElements = scenes[currentScene];
    const elementIndex = sceneElements.indexOf(menuEl);
    if (elementIndex > -1) {
      sceneElements.splice(elementIndex, 1);
    }
  });
}




function getFontSize(fontSize) {
    const sizes = {
        title: titleSizeInput.value,
        big: bigSizeInput.value,
        medium: mediumSizeInput.value,
        small: smallSizeInput.value
    };
    return sizes[fontSize] || 24; // Default to medium size if undefined
}

function setBackground() {
    const file = backgroundFileInput.files[0];
    if (file) {
        backgroundPath = file.name;
        const reader = new FileReader();
        reader.onload = function(event) {
            canvas.style.backgroundImage = `url(${event.target.result})`;
        };
        reader.readAsDataURL(file);
    }
}

function createJukaApp() {
    updateScenes();
const elements = Array.from(canvas.getElementsByClassName('element'));
  const config = {
    title: document.getElementById('title').value,
    author: document.getElementById('author').value,
    description: document.getElementById('description').value,
    variables: {
      ...variables,
      buttonColor: { r: 255, g: 0, b: 0 },
      labelColor: { r: 255, g: 255, b: 255 },
      backgroundImage: backgroundPath,
      fonts: {
        title: 'Roboto-Black.ttf',
        big: 'Roboto-Black.ttf',
        medium: 'Roboto-Black.ttf',
        small: 'Roboto-Black.ttf'
      },
      fontSizes: {
        title: parseInt(titleSizeInput.value, 10),
        big: parseInt(bigSizeInput.value, 10),
        medium: parseInt(mediumSizeInput.value, 10),
        small: parseInt(smallSizeInput.value, 10)
      }
    },
    scenes: Object.keys(scenes).map(sceneName => ({
      name: sceneName,
      background: sceneName === currentScene ? backgroundPath : '',
      elements: scenes[sceneName].map(el => {
        const elementType = el.getAttribute('data-type');
        const baseElement = {
          type: elementType,
          x: parseInt(el.getAttribute('data-x'), 10),
          y: parseInt(el.getAttribute('data-y'), 10)
        };
        
        if (elementType === 'menu') {
          return {
            ...baseElement,
            height: parseInt(el.getAttribute('data-height'), 10) || 50
          };
        } else {
          return {
            ...baseElement,
            text: el.querySelector('.text-content')?.textContent || '',
            color: el.getAttribute('data-color'),
            font: el.getAttribute('data-font'),
            bgColor: el.getAttribute('data-type') === 'button' ? el.getAttribute('data-bg-color') : undefined,
            trigger: el.getAttribute('data-trigger'),
            triggerTarget: el.getAttribute('data-trigger-target'),
            triggerValue: el.getAttribute('data-trigger-value'),
            image: el.getAttribute('data-image'),
            width: el.getAttribute('data-width'),
            height: el.getAttribute('data-height')
          };
        }
      })
    }))
  };

  const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(config, null, 2));
  const downloadAnchorNode = document.createElement('a');
  downloadAnchorNode.setAttribute("href", dataStr);
  downloadAnchorNode.setAttribute("download", "jukaconfig.json");
  document.body.appendChild(downloadAnchorNode);
  downloadAnchorNode.click();
  downloadAnchorNode.remove();
};


canvas.addEventListener('click', (event) => {
    if (event.target === canvas) {
        hideControls();
    }
});

function hideControls() {
    document.querySelector('.control-label[for="datax"]').style.display = 'none';
    datax.style.display = 'none';
    document.querySelector('.control-label[for="datay"]').style.display = 'none';
    datay.style.display = 'none';
    document.querySelector('.control-label[for="colorPicker"]').style.display = 'none';
    colorPicker.style.display = 'none';
    document.querySelector('.control-label[for="bgColorPicker"]').style.display = 'none';
    bgColorPicker.style.display = 'none';
    document.querySelector('.control-label[for="fontSizePicker"]').style.display = 'none';
    fontSizePicker.style.display = 'none';
    document.querySelector('.control-label[for="triggerSelector"]').style.display = 'none';
    triggerSelector.style.display = 'none';
    sceneChangeSelector.style.display = 'none';
    externalAppPath.style.display = 'none';
    variableChangeSelector.style.display = 'none';
    variableChangeValue.style.display = 'none';
}




document.getElementById('loadFile').addEventListener('change', function(event) {
    const file = event.target.files[0];
    if (file) {
        const reader = new FileReader();
        reader.onload = function(event) {
            const data = JSON.parse(event.target.result);
            loadJukaApp(data);
        };
        reader.readAsText(file);
    }
});

function loadJukaApp(data) {
    // Update title, author, description
    document.getElementById('title').value = data.title || '';
    document.getElementById('author').value = data.author || '';
    document.getElementById('description').value = data.description || '';

    // Update font sizes
    if (data.variables && data.variables.fontSizes) {
        document.getElementById('titleSize').value = data.variables.fontSizes.title || 48;
        document.getElementById('bigSize').value = data.variables.fontSizes.big || 36;
        document.getElementById('mediumSize').value = data.variables.fontSizes.medium || 24;
        document.getElementById('smallSize').value = data.variables.fontSizes.small || 18;
    }

    // Clear the scenes object
    scenes = {};

    // Clear the canvas and scene selector
    canvas.innerHTML = '';
    const sceneSelector = document.getElementById('sceneSelector');
    sceneSelector.innerHTML = '';

    // Load variables if they exist
    variables = {};
    if (data.variables) {
        // Extract any custom variables (excluding predefined ones)
        Object.keys(data.variables).forEach(key => {
            if (!['buttonColor', 'labelColor', 'backgroundImage', 'fonts', 'fontSizes'].includes(key)) {
                variables[key] = data.variables[key];
                
                // Update variable selectors
                const option = document.createElement('option');
                option.value = key;
                option.textContent = key;
                variableSelector.appendChild(option);
                
                const variableOption = document.createElement('option');
                variableOption.value = key;
                variableOption.textContent = key;
                variableChangeSelector.appendChild(variableOption);
            }
        });
    }

    // Load scenes and elements
    data.scenes.forEach(scene => {
        // Create a new scene entry
        scenes[scene.name] = [];
        
        // Process each element in the scene
scene.elements.forEach(element => {
    let el;
    
    // Handle menu elements differently
    if (element.type === 'menu') {
        el = document.createElement('div');
        el.classList.add('element', 'menu');
        el.style.position = 'absolute';
        el.style.left = '0px';  // Force to 0px horizontal position
        el.style.top = `${element.y}px`;
        el.style.width = '100%';
        el.style.height = `${element.height || 50}px`;
        
        // Create menu structure
        el.innerHTML = `
      <div class="handle"></div>
      <div class="menu-scene-buttons"></div>
      <button class="menu-language">EN</button>
      <div class="menu-clock">00:00</div>
        `;
        
        // Set attribute for height
        el.setAttribute('data-height', element.height || 50);
        el.setAttribute('data-type', element.type);
        el.setAttribute('data-x', 0);  // Set to 0 regardless of saved value
        el.setAttribute('data-y', element.y);
        
        // Initialize the menu
        setupMenuEvents(el);
        initializeMenu(el);
        
        // Start the clock
        updateMenuClock(el.querySelector('.menu-clock'));
    } else {
        // Create a regular element (button, label, image)
        el = document.createElement('div');
        el.classList.add('element');
        el.style.position = 'absolute';
        el.style.left = `${element.x}px`;
        el.style.top = `${element.y}px`;
        
        // Add the text content span
        const textSpan = document.createElement('span');
        textSpan.classList.add('text-content');
        textSpan.textContent = element.text || element.type.charAt(0).toUpperCase() + element.type.slice(1);
        el.appendChild(textSpan);
        
        // Add the remove button
        const removeButton = document.createElement('span');
        removeButton.textContent = '✕';
        removeButton.classList.add('remove-button');
        el.appendChild(removeButton);
        
        // Set attributes
        el.setAttribute('data-type', element.type);
        el.setAttribute('data-x', element.x);
        el.setAttribute('data-y', element.y);
        
        // Apply specific styling based on element type
        if (element.type === 'button' || element.type === 'label') {
            el.setAttribute('data-color', element.color || '#000000');
            el.style.color = element.color || '#000000';
            el.setAttribute('data-font', element.font || 'medium');
            el.style.fontSize = `${getFontSize(element.font || 'medium')}px`;
    
            if (element.type === 'button') {
                el.setAttribute('data-bg-color', element.bgColor || '#ffffff');
                el.style.backgroundColor = element.bgColor || '#ffffff';
                el.setAttribute('data-trigger', element.trigger || '');
                el.setAttribute('data-trigger-target', element.triggerTarget || '');
                el.setAttribute('data-trigger-value', element.triggerValue || '');
            } else if (element.type === 'label') {
                // Ensure labels have transparent background
                el.style.background = 'none';
            }
        } else if (element.type === 'image') {
            if (element.image) {
                el.style.backgroundImage = `url(${element.image})`;
                el.style.backgroundSize = 'cover';
                el.setAttribute('data-image', element.image);
            }
            
            el.style.width = `${element.width || 100}px`;
            el.style.height = `${element.height || 100}px`;
            el.setAttribute('data-width', element.width || 100);
            el.setAttribute('data-height', element.height || 100);
        }
        
        // Set up the event listeners
        setupElementEvents(el);
    }
    
    // Add the element to the canvas
    canvas.appendChild(el);
    
    // Add to the scene array (store the cloned DOM element)
    scenes[scene.name].push(el.cloneNode(true));
});

        // Add scene to the scene selector
        const option = document.createElement('option');
        option.value = scene.name;
        option.textContent = scene.name;
        sceneSelector.appendChild(option);
    });

    // Set the background if available
    if (data.scenes.length > 0) {
        const firstScene = data.scenes[0];
        if (firstScene.background) {
            canvas.style.backgroundImage = `url(${firstScene.background})`;
            backgroundPath = firstScene.background;
        }
    }

    // Show the first scene by default
    if (data.scenes.length > 0) {
        currentScene = data.scenes[0].name;
        sceneSelector.value = currentScene;
    }

    // Update scene change selector with all available scenes
    updateSceneChangeSelector();
    
    // Apply font sizes to all elements
    updateElementFontSizes();
    
    // Update any variable references in text
    updateVariableText();
  document.querySelectorAll('.menu').forEach(menuEl => {
    initializeMenu(menuEl);
  });
}


// Automatically load jukaconfig.json
window.addEventListener('load', () => {
    fetch('player/jukaconfig.json')
        .then(response => response.json())
        .then(data => loadJukaApp(data))
        .catch(error => console.error('Error loading jukaconfig.json:', error));

    customWidthInput.addEventListener('change', updateCanvasSize);
    customHeightInput.addEventListener('change', updateCanvasSize);
});

// Add a button to clear everything and start new
document.getElementById('clearButton').addEventListener('click', () => {
    if (confirm('Are you sure you want to clear everything and start new?')) {
        scenes = { 'Scene 1': [] };
        currentScene = 'Scene 1';
        variables = {};
        canvas.innerHTML = '';
        sceneSelector.innerHTML = '';
        const option = document.createElement('option');
        option.value = 'Scene 1';
        option.textContent = 'Scene 1';
        sceneSelector.appendChild(option);
        sceneSelector.value = 'Scene 1';
        updateSceneChangeSelector();
        updateCanvasSize();
        document.getElementById('title').value = '';
        document.getElementById('author').value = '';
        document.getElementById('description').value = '';
        document.getElementById('titleSize').value = 48;
        document.getElementById('bigSize').value = 36;
        document.getElementById('mediumSize').value = 24;
        document.getElementById('smallSize').value = 18;
        // Add menu to new scene
        const canvasHeight = canvas.offsetHeight;
        addElement('menu', 0, canvasHeight - 50);
        
        document.querySelectorAll('.menu').forEach(menuEl => {
            updateMenuSceneButtons(menuEl);
        });
    }
});


// Add a dialogue when closing the tab
window.addEventListener('beforeunload', (event) => {
    event.preventDefault();
    event.returnValue = 'Are you sure you want to leave? Changes you made may not be saved.';
});

