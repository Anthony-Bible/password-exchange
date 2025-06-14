// Helper function for unbiased random selection using rejection sampling
function getUnbiasedRandomIndex(max) {
    const maxValidValue = Math.floor(0xFFFFFFFF / max) * max;
    let randomValue;
    
    do {
        const array = new Uint32Array(1);
        crypto.getRandomValues(array);
        randomValue = array[0];
    } while (randomValue >= maxValidValue);
    
    return randomValue % max;
}

// Secure Fisher-Yates shuffle using crypto.getRandomValues
function secureShuffleArray(array) {
    const shuffled = [...array]; // Create a copy to avoid mutating original
    
    for (let i = shuffled.length - 1; i > 0; i--) {
        const j = getUnbiasedRandomIndex(i + 1);
        [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
    }
    
    return shuffled;
}

// Global variables for EFF word list and error state
let effWordList = [];
let wordListStatus = 'loading'; // 'loading', 'loaded', 'error', 'fallback'
let wordListRetryCount = 0;
const MAX_RETRY_ATTEMPTS = 3;

// Show loading indicator for wordlist
function showWordlistStatus(status, message = '') {
    const statusElements = document.querySelectorAll('.wordlist-status');
    statusElements.forEach(el => {
        switch(status) {
            case 'loading':
                el.innerHTML = '<small class="text-info"><i class="fas fa-spinner fa-spin me-1"></i>Loading word list...</small>';
                break;
            case 'error':
                el.innerHTML = '<small class="text-warning"><i class="fas fa-exclamation-triangle me-1"></i>Using limited word list</small>';
                break;
            case 'loaded':
                el.innerHTML = '<small class="text-success"><i class="fas fa-check me-1"></i>Enhanced word list loaded</small>';
                break;
            case 'fallback':
                el.innerHTML = '<small class="text-warning"><i class="fas fa-info-circle me-1"></i>Using fallback word list</small>';
                break;
            default:
                el.innerHTML = '';
        }
    });
}

// Download EFF word list with retry mechanism
async function loadEFFWordList() {
    wordListStatus = 'loading';
    showWordlistStatus('loading');
    
    try {
        const response = await fetch('./eff_large_wordlist.txt');
        if (!response.ok) {
            throw new Error(`Failed to fetch wordlist: ${response.status} ${response.statusText}`);
        }
        const text = await response.text();
        
        if (!text || text.trim().length === 0) {
            throw new Error('Wordlist file is empty');
        }
        
        // Parse the word list (format is "number\tword")
        const parsedWords = text.trim().split('\n')
            .map(line => {
                const parts = line.split('\t');
                return parts[1]; // Return just the word part
            })
            .filter(word => word && word.length > 0);
        
        if (parsedWords.length < 100) {
            throw new Error(`Wordlist too small: only ${parsedWords.length} words found`);
        }
        
        effWordList = parsedWords;
        wordListStatus = 'loaded';
        showWordlistStatus('loaded');
        console.log(`Successfully loaded ${effWordList.length} words from EFF word list`);
        
    } catch (error) {
        console.warn('Failed to load EFF word list:', error.message);
        
        // Retry logic
        if (wordListRetryCount < MAX_RETRY_ATTEMPTS) {
            wordListRetryCount++;
            console.log(`Retrying wordlist load (attempt ${wordListRetryCount}/${MAX_RETRY_ATTEMPTS})...`);
            setTimeout(() => loadEFFWordList(), 1000 * wordListRetryCount); // Exponential backoff
            return;
        }
        
        // Use fallback word list after all retries failed
        wordListStatus = 'fallback';
        showWordlistStatus('fallback');
        effWordList = [
            'adventure', 'airplane', 'alphabet', 'amazing', 'animal', 'awesome', 'balance', 'banana',
            'beautiful', 'bicycle', 'birthday', 'butterfly', 'calendar', 'camera', 'celebrate',
            'challenge', 'champion', 'chocolate', 'computer', 'creative', 'delicious', 'diamond',
            'dinosaur', 'elephant', 'energy', 'exciting', 'fantastic', 'favorite', 'freedom',
            'friendship', 'garden', 'gigantic', 'grateful', 'happiness', 'harmony', 'helicopter',
            'imagination', 'incredible', 'journey', 'keyboard', 'laughter', 'learning', 'library',
            'lightning', 'magical', 'mountain', 'musical', 'mystery', 'nature', 'notebook',
            'ocean', 'opportunity', 'optimistic', 'painting', 'paradise', 'password', 'peaceful',
            'penguin', 'perfect', 'playground', 'positive', 'powerful', 'princess', 'progress',
            'rainbow', 'remember', 'satellite', 'scientist', 'security', 'solution', 'spaceship',
            'special', 'success', 'sunshine', 'surprise', 'technology', 'telescope', 'treasure',
            'umbrella', 'universe', 'vacation', 'victory', 'volcano', 'waterfall', 'wonderful'
        ];
        
        console.log(`Using fallback word list with ${effWordList.length} words`);
    }
}

// Generate passphrase using EFF word list
function generatePassphrase(wordCount, includeNumbers, includeSymbols, separator = '-') {
    if (effWordList.length === 0) {
        throw new Error('Word list is still loading. Please wait a moment and try again.');
    }
    
    if (wordCount < 1 || wordCount > 12) {
        throw new Error('Word count must be between 1 and 12 words.');
    }
    
    const words = [];
    const numbers = '0123456789';
    const symbols = '!@#$%^&*';
    
    // Select random words
    for (let i = 0; i < wordCount; i++) {
        const randomIndex = getUnbiasedRandomIndex(effWordList.length);
        words.push(effWordList[randomIndex]);
    }
    
    // Join words with specified separator
    let passphrase = words.join(separator);
    
    // Optionally add numbers
    if (includeNumbers) {
        const numCount = Math.floor(Math.random() * 3) + 1; // 1-3 numbers
        for (let i = 0; i < numCount; i++) {
            passphrase += numbers[getUnbiasedRandomIndex(numbers.length)];
        }
    }
    
    // Optionally add symbols
    if (includeSymbols) {
        const symCount = Math.floor(Math.random() * 2) + 1; // 1-2 symbols
        for (let i = 0; i < symCount; i++) {
            passphrase += symbols[getUnbiasedRandomIndex(symbols.length)];
        }
    }
    
    return passphrase;
}

// Calculate password strength using zxcvbn
function analyzePassword(password) {
    const length = password.length;
    let hasLower = false, hasUpper = false, hasNumbers = false, hasSymbols = false;
    
    // Still check character types for display purposes
    for (const ch of password) {
        if (/[a-z]/.test(ch)) hasLower = true;
        if (/[A-Z]/.test(ch)) hasUpper = true;
        if (/[0-9]/.test(ch)) hasNumbers = true;
        if (/[^a-zA-Z0-9]/.test(ch)) hasSymbols = true;
    }
    
    // Use zxcvbn for realistic password strength assessment
    const result = zxcvbn(password);
    
    // Map zxcvbn score (0-4) to our strength labels and scores
    const strengthMapping = {
        0: { strength: 'Very Weak', strengthScore: 10, strengthClass: 'bg-danger' },
        1: { strength: 'Weak', strengthScore: 25, strengthClass: 'bg-danger' },
        2: { strength: 'Fair', strengthScore: 50, strengthClass: 'bg-warning' },
        3: { strength: 'Good', strengthScore: 75, strengthClass: 'bg-info' },
        4: { strength: 'Strong', strengthScore: 100, strengthClass: 'bg-success' }
    };
    
    const { strength, strengthScore, strengthClass } = strengthMapping[result.score];
    
    return {
        entropy: Math.round(result.guesses_log10 * 3.32), // Convert log10 to bits (log2)
        strength,
        strengthScore,
        strengthClass,
        hasLower,
        hasUpper,
        hasNumbers,
        hasSymbols,
        length,
        crackTime: result.crack_times_display.offline_slow_hashing_1e4_per_second,
        feedback: result.feedback,
        zxcvbnResult: result
    };
}

// Update password strength display
function updatePasswordStrength(password) {
    const analysis = analyzePassword(password);
    
    // Update strength bar
    const strengthBar = document.getElementById('strength-bar');
    const strengthText = document.getElementById('strength-text');
    const entropyText = document.getElementById('entropy-text');
    
    strengthBar.style.width = analysis.strengthScore + '%';
    strengthBar.className = `progress-bar ${analysis.strengthClass}`;
    strengthBar.setAttribute('aria-valuenow', analysis.strengthScore);
    
    strengthText.textContent = analysis.strength;
    strengthText.className = `text-muted ${analysis.strengthClass.replace('bg-', 'text-')}`;
    
    // Show time to crack instead of just entropy
    entropyText.textContent = `Time to crack: ${analysis.crackTime}`;
    
    // Update character types
    const charTypesDiv = document.getElementById('char-types');
    const types = [];
    if (analysis.hasLower) types.push('lowercase');
    if (analysis.hasUpper) types.push('UPPERCASE');
    if (analysis.hasNumbers) types.push('123');
    if (analysis.hasSymbols) types.push('!@#');
    
    charTypesDiv.innerHTML = `<i class="fas fa-info-circle me-1"></i>Contains: ${types.join(', ')}`;
    
    // Update security features with zxcvbn feedback
    const securityDiv = document.getElementById('security-features');
    const features = [];
    if (analysis.length >= 12) features.push('Long');
    if (analysis.zxcvbnResult.score >= 3) features.push('Strong pattern');
    if (analysis.entropy >= 60) features.push('High entropy');
    
    securityDiv.innerHTML = `<i class="fas fa-shield-alt me-1"></i>${features.length ? features.join(', ') : 'Basic'}`;
    
    // Remove any existing warnings and suggestions
    const existingFeedback = document.querySelectorAll('#password-analysis .alert');
    existingFeedback.forEach(element => element.remove());
    
    // Add zxcvbn warnings if present
    if (analysis.feedback.warning) {
        const warningDiv = document.createElement('div');
        warningDiv.className = 'alert alert-warning mt-2 small';
        warningDiv.innerHTML = `
            <i class="fas fa-exclamation-triangle me-1"></i>
            <strong>Warning:</strong> ${analysis.feedback.warning}
        `;
        document.getElementById('password-analysis').appendChild(warningDiv);
    }
    
    // Add zxcvbn feedback if there are suggestions
    if (analysis.feedback.suggestions && analysis.feedback.suggestions.length > 0) {
        const feedbackDiv = document.createElement('div');
        feedbackDiv.className = 'alert alert-info mt-2 small';
        feedbackDiv.innerHTML = `
            <i class="fas fa-lightbulb me-1"></i>
            <strong>Suggestions:</strong> ${analysis.feedback.suggestions.join(' ')}
        `;
        document.getElementById('password-analysis').appendChild(feedbackDiv);
    }
}

// Secure password generator using crypto.getRandomValues with rejection sampling
// Ensures at least one character from each selected character class
function generateSecurePassword(length, includeUppercase, includeLowercase, includeNumbers, includeSymbols, excludeAmbiguous = false) {
    let uppercase = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    let lowercase = 'abcdefghijklmnopqrstuvwxyz';
    let numbers = '0123456789';
    const symbols = '!@#$%^&*()_+-=[]{}|;:,.<>?';
    
    // Remove ambiguous characters if requested
    if (excludeAmbiguous) {
        uppercase = uppercase.replace(/[OI]/g, ''); // Remove O and I
        lowercase = lowercase.replace(/[l]/g, ''); // Remove l
        numbers = numbers.replace(/[01]/g, ''); // Remove 0 and 1
    }
    
    // Build pools of selected character types
    const pools = [];
    if (includeUppercase) pools.push(uppercase);
    if (includeLowercase) pools.push(lowercase);
    if (includeNumbers) pools.push(numbers);
    if (includeSymbols) pools.push(symbols);
    
    if (!pools.length) throw new Error("Select at least one character set");
    
    if (length < pools.length) {
        throw new Error(`Password length must be at least ${pools.length} to include all selected character types`);
    }
    
    // 1️⃣ ensure one from each pool
    const bytes = crypto.getRandomValues(new Uint8Array(length));
    const pwd = pools.map(p => p[getUnbiasedRandomIndex(p.length)]);
    
    // 2️⃣ fill the rest from the union
    const union = pools.join("");
    for (let i = pwd.length; i < length; i++) {
        pwd.push(union[getUnbiasedRandomIndex(union.length)]);
    }
    
    // 3️⃣ shuffle to avoid predictable prefix
    const shuffled = secureShuffleArray(pwd);
    return shuffled.join("");
}

// Helper function to convert technical error messages to user-friendly ones
function getUserFriendlyErrorMessage(errorMessage) {
    if (errorMessage.includes('Word list')) {
        return 'The word dictionary is still loading. Please wait a moment and try again.';
    }
    if (errorMessage.includes('character')) {
        return 'Please select at least one character type (uppercase, lowercase, numbers, or symbols).';
    }
    if (errorMessage.includes('length')) {
        return 'Password length requirements cannot be met with current settings. Try adjusting the length or character options.';
    }
    if (errorMessage.includes('crypto') || errorMessage.includes('random')) {
        return 'Unable to generate secure random values. Please try refreshing the page.';
    }
    if (errorMessage.includes('fetch') || errorMessage.includes('network')) {
        return 'Network connection issue. Using offline word list instead.';
    }
    return 'An unexpected error occurred. Please try again or refresh the page.';
}

// Initialize password generator functionality
function initializePasswordGenerator() {
    // Load EFF word list on page load
    loadEFFWordList();
    
    // Password generator modal functionality
    const modalLengthSlider = document.getElementById('modal-password-length');
    const modalLengthValue = document.getElementById('modalLengthValue');
    const modalGenerateButton = document.getElementById('modal-generate-password');
    const generatedPasswordSection = document.getElementById('generated-password-section');
    const generatedPasswordInput = document.getElementById('generated-password');
    const copyPasswordBtn = document.getElementById('copy-password-btn');
    const insertPasswordBtn = document.getElementById('insert-password-btn');
    const messageTextarea = document.getElementById('form_message');
    
    // Update length display in modal and trigger real-time generation
    modalLengthSlider.addEventListener('input', function() {
        modalLengthValue.textContent = this.value;
        generatePasswordRealtime();
    });
    
    // Update passphrase word count display and trigger real-time generation
    const passphraseWordCountSlider = document.getElementById('passphrase-word-count');
    const passphraseWordCountValue = document.getElementById('passphraseWordCountValue');
    
    passphraseWordCountSlider.addEventListener('input', function() {
        passphraseWordCountValue.textContent = this.value;
        generatePasswordRealtime();
    });
    
    // Password type toggle functionality with real-time generation
    const passwordTypeRadios = document.querySelectorAll('input[name="passwordType"]');
    const passphraseOptions = document.getElementById('passphrase-options');
    
    passwordTypeRadios.forEach(radio => {
        radio.addEventListener('change', function() {
            if (this.value === 'passphrase') {
                passphraseOptions.style.display = 'block';
            } else {
                passphraseOptions.style.display = 'none';
            }
            generatePasswordRealtime();
        });
    });
    
    // Separator selection functionality with real-time generation
    const separatorSelect = document.getElementById('passphrase-separator');
    const customSeparatorInput = document.getElementById('custom-separator');
    
    separatorSelect.addEventListener('change', function() {
        if (this.value === 'custom') {
            customSeparatorInput.style.display = 'block';
            customSeparatorInput.focus();
        } else {
            customSeparatorInput.style.display = 'none';
        }
        generatePasswordRealtime();
    });
    
    // Custom separator input with real-time generation
    customSeparatorInput.addEventListener('input', function() {
        generatePasswordRealtime();
    });
    
    // Add real-time generation to all character type checkboxes
    const characterTypeCheckboxes = [
        document.getElementById('modal-include-uppercase'),
        document.getElementById('modal-include-lowercase'),
        document.getElementById('modal-include-numbers'),
        document.getElementById('modal-include-symbols'),
        document.getElementById('exclude-ambiguous'),
        document.getElementById('include-numbers-passphrase'),
        document.getElementById('include-symbols-passphrase')
    ];
    
    characterTypeCheckboxes.forEach(checkbox => {
        checkbox.addEventListener('change', function() {
            generatePasswordRealtime();
        });
    });
    
    // Helper function to get selected separator
    function getSelectedSeparator() {
        const separatorValue = separatorSelect.value;
        if (separatorValue === 'custom') {
            return customSeparatorInput.value || '-'; // Default to dash if custom is empty
        }
        return separatorValue;
    }
    
    // Real-time password generation function
    let realtimeGenerationTimeout;
    function generatePasswordRealtime() {
        // Clear existing timeout to debounce rapid changes
        clearTimeout(realtimeGenerationTimeout);
        
        realtimeGenerationTimeout = setTimeout(() => {
            try {
                const length = parseInt(modalLengthSlider.value);
                const includeUppercase = document.getElementById('modal-include-uppercase').checked;
                const includeLowercase = document.getElementById('modal-include-lowercase').checked;
                const includeNumbers = document.getElementById('modal-include-numbers').checked;
                const includeSymbols = document.getElementById('modal-include-symbols').checked;
                const excludeAmbiguous = document.getElementById('exclude-ambiguous').checked;
                const passwordType = document.querySelector('input[name="passwordType"]:checked').value;
                
                let password;
                
                if (passwordType === 'passphrase') {
                    const wordCount = parseInt(document.getElementById('passphrase-word-count').value);
                    const includeNumbersPassphrase = document.getElementById('include-numbers-passphrase').checked;
                    const includeSymbolsPassphrase = document.getElementById('include-symbols-passphrase').checked;
                    const separator = getSelectedSeparator();
                    password = generatePassphrase(wordCount, includeNumbersPassphrase, includeSymbolsPassphrase, separator);
                } else {
                    if (!includeUppercase && !includeLowercase && !includeNumbers && !includeSymbols) {
                        // Show empty state when no character types selected
                        generatedPasswordInput.value = '';
                        generatedPasswordSection.style.display = 'block';
                        insertPasswordBtn.style.display = 'none';
                        document.getElementById('strength-text').textContent = 'Select at least one character type';
                        document.getElementById('entropy-text').textContent = '';
                        document.getElementById('char-types').innerHTML = '';
                        document.getElementById('security-features').innerHTML = '';
                        return;
                    }
                    password = generateSecurePassword(length, includeUppercase, includeLowercase, includeNumbers, includeSymbols, excludeAmbiguous);
                }
                
                generatedPasswordInput.value = password;
                generatedPasswordSection.style.display = 'block';
                insertPasswordBtn.style.display = 'inline-block';
                
                // Hide multiple options section if visible
                document.getElementById('password-options-section').style.display = 'none';
                
                // Update password strength analysis
                updatePasswordStrength(password);
                
            } catch (error) {
                console.error('Real-time password generation error:', error);
                
                // Show user-friendly error message
                generatedPasswordInput.value = '';
                insertPasswordBtn.style.display = 'none';
                
                let errorMessage = 'Unable to generate password';
                if (error.message.includes('Word list')) {
                    errorMessage = 'Word list loading... Please wait';
                } else if (error.message.includes('character')) {
                    errorMessage = 'Select at least one character type';
                } else if (error.message.includes('length')) {
                    errorMessage = 'Password length requirements not met';
                }
                
                document.getElementById('strength-text').textContent = errorMessage;
                document.getElementById('entropy-text').textContent = '';
                document.getElementById('char-types').innerHTML = '';
                document.getElementById('security-features').innerHTML = '';
            }
        }, 200); // 200ms debounce delay
    }
    
    // Generate password button in modal (now triggers manual generation)
    modalGenerateButton.addEventListener('click', function() {
        // Generate a new password immediately without debounce
        clearTimeout(realtimeGenerationTimeout);
        generatePasswordRealtime();
        
        // Visual feedback
        const originalText = this.innerHTML;
        this.innerHTML = '<i class="fas fa-sync-alt me-2"></i>Regenerated!';
        this.classList.remove('btn-primary');
        this.classList.add('btn-success');
        
        setTimeout(() => {
            this.innerHTML = originalText;
            this.classList.remove('btn-success');
            this.classList.add('btn-primary');
        }, 1000);
    });
    
    // Copy password to clipboard
    copyPasswordBtn.addEventListener('click', async function() {
        try {
            await navigator.clipboard.writeText(generatedPasswordInput.value);
            
            // Visual feedback
            const originalHTML = this.innerHTML;
            this.innerHTML = '<i class="fas fa-check"></i>';
            this.classList.remove('btn-outline-secondary');
            this.classList.add('btn-success');
            
            setTimeout(() => {
                this.innerHTML = originalHTML;
                this.classList.remove('btn-success');
                this.classList.add('btn-outline-secondary');
            }, 1500);
            
        } catch (error) {
            console.error('Copy failed:', error);
            
            try {
                // Fallback for older browsers
                generatedPasswordInput.select();
                const success = document.execCommand('copy');
                
                if (success) {
                    // Visual feedback for fallback copy
                    const originalHTML = this.innerHTML;
                    this.innerHTML = '<i class="fas fa-check"></i>';
                    this.classList.remove('btn-outline-secondary');
                    this.classList.add('btn-success');
                    
                    setTimeout(() => {
                        this.innerHTML = originalHTML;
                        this.classList.remove('btn-success');
                        this.classList.add('btn-outline-secondary');
                    }, 1500);
                } else {
                    throw new Error('Fallback copy also failed');
                }
            } catch (fallbackError) {
                // If both modern and fallback methods fail
                const originalHTML = this.innerHTML;
                this.innerHTML = '<i class="fas fa-times"></i>';
                this.classList.remove('btn-outline-secondary');
                this.classList.add('btn-danger');
                
                setTimeout(() => {
                    this.innerHTML = originalHTML;
                    this.classList.remove('btn-danger');
                    this.classList.add('btn-outline-secondary');
                }, 2000);
                
                // Show error message in the password input temporarily
                const originalValue = generatedPasswordInput.value;
                generatedPasswordInput.value = 'Copy failed - please select and copy manually';
                generatedPasswordInput.select();
                
                setTimeout(() => {
                    generatedPasswordInput.value = originalValue;
                }, 3000);
            }
        }
    });
    
    // Insert password into message textarea
    insertPasswordBtn.addEventListener('click', function() {
        messageTextarea.value = generatedPasswordInput.value;
        
        // Close the modal
        const modal = bootstrap.Modal.getInstance(document.getElementById('passwordGeneratorModal'));
        modal.hide();
        
        // Focus on the textarea
        messageTextarea.focus();
    });
    
    // Generate multiple password options
    const modalGenerateMultiple = document.getElementById('modal-generate-multiple');
    const passwordOptionsSection = document.getElementById('password-options-section');
    const passwordOptionsList = document.getElementById('password-options-list');
    
    modalGenerateMultiple.addEventListener('click', function() {
        try {
            const length = parseInt(modalLengthSlider.value);
            const includeUppercase = document.getElementById('modal-include-uppercase').checked;
            const includeLowercase = document.getElementById('modal-include-lowercase').checked;
            const includeNumbers = document.getElementById('modal-include-numbers').checked;
            const includeSymbols = document.getElementById('modal-include-symbols').checked;
            const excludeAmbiguous = document.getElementById('exclude-ambiguous').checked;
            const passwordType = document.querySelector('input[name="passwordType"]:checked').value;
            
            // Generate 5 options
            const passwords = [];
            for (let i = 0; i < 5; i++) {
                let password;
                if (passwordType === 'passphrase') {
                    const wordCount = parseInt(document.getElementById('passphrase-word-count').value);
                    const includeNumbersPassphrase = document.getElementById('include-numbers-passphrase').checked;
                    const includeSymbolsPassphrase = document.getElementById('include-symbols-passphrase').checked;
                    const separator = getSelectedSeparator();
                    password = generatePassphrase(wordCount, includeNumbersPassphrase, includeSymbolsPassphrase, separator);
                } else {
                    if (!includeUppercase && !includeLowercase && !includeNumbers && !includeSymbols) {
                        alert('Please select at least one character type for password generation.');
                        return;
                    }
                    password = generateSecurePassword(length, includeUppercase, includeLowercase, includeNumbers, includeSymbols, excludeAmbiguous);
                }
                passwords.push(password);
            }
            
            // Clear previous options
            passwordOptionsList.innerHTML = '';
            
            // Create clickable options
            passwords.forEach((password, index) => {
                const analysis = analyzePassword(password);
                const optionElement = document.createElement('button');
                optionElement.className = 'list-group-item list-group-item-action d-flex justify-content-between align-items-center';
                optionElement.innerHTML = `
                    <div class="d-flex flex-column align-items-start">
                        <code class="font-monospace">${password}</code>
                        <small class="text-muted">${analysis.strength} (${analysis.entropy} bits)</small>
                    </div>
                    <i class="fas fa-arrow-right text-primary"></i>
                `;
                
                optionElement.addEventListener('click', function() {
                    generatedPasswordInput.value = password;
                    generatedPasswordSection.style.display = 'block';
                    insertPasswordBtn.style.display = 'inline-block';
                    passwordOptionsSection.style.display = 'none';
                    updatePasswordStrength(password);
                });
                
                passwordOptionsList.appendChild(optionElement);
            });
            
            // Show options section, hide single password section
            passwordOptionsSection.style.display = 'block';
            generatedPasswordSection.style.display = 'none';
            
            // Visual feedback
            const originalText = this.innerHTML;
            this.innerHTML = '<i class="fas fa-check me-2"></i>Generated!';
            this.classList.remove('btn-outline-info');
            this.classList.add('btn-success');
            
            setTimeout(() => {
                this.innerHTML = originalText;
                this.classList.remove('btn-success');
                this.classList.add('btn-outline-info');
            }, 1500);
            
        } catch (error) {
            console.error('Password options generation error:', error);
            
            // Show user-friendly error in the UI instead of alert
            passwordOptionsList.innerHTML = `
                <div class="list-group-item text-center text-danger">
                    <i class="fas fa-exclamation-triangle me-2"></i>
                    <strong>Generation Error:</strong><br>
                    <small>${getUserFriendlyErrorMessage(error.message)}</small>
                    <br><br>
                    <button class="btn btn-sm btn-outline-primary" onclick="this.closest('.list-group-item').remove(); document.getElementById('modal-generate-multiple').click();">
                        <i class="fas fa-retry me-1"></i>Try Again
                    </button>
                </div>
            `;
            passwordOptionsSection.style.display = 'block';
        }
    });
    
    // Initialize password generation when modal opens
    document.getElementById('passwordGeneratorModal').addEventListener('shown.bs.modal', function() {
        // Generate initial password when modal opens
        generatePasswordRealtime();
    });
    
    // Reset modal when it's closed
    document.getElementById('passwordGeneratorModal').addEventListener('hidden.bs.modal', function() {
        generatedPasswordSection.style.display = 'none';
        passwordOptionsSection.style.display = 'none';
        insertPasswordBtn.style.display = 'none';
        generatedPasswordInput.value = '';
        passwordOptionsList.innerHTML = '';
        
        // Clear any pending generation timeouts
        clearTimeout(realtimeGenerationTimeout);
    });
}