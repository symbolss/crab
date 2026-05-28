package com.family.crab.child.ui.pairing

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Toast
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.navigation.fragment.findNavController
import com.family.crab.child.R
import com.family.crab.child.databinding.FragmentPairingBinding
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class PairingFragment : Fragment() {
    
    private var _binding: FragmentPairingBinding? = null
    private val binding get() = _binding!!
    
    private val viewModel: PairingViewModel by viewModels()
    
    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = FragmentPairingBinding.inflate(inflater, container, false)
        return binding.root
    }
    
    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        setupObservers()
        setupListeners()
    }
    
    private fun setupListeners() {
        binding.btnPair.setOnClickListener {
            val code = binding.etPairingCode.text.toString()
            viewModel.pairDevice(code)
        }
    }
    
    private fun setupObservers() {
        viewModel.pairingState.observe(viewLifecycleOwner) { state ->
            when (state) {
                is PairingViewModel.PairingState.Idle -> {
                    showLoading(false)
                    showError(null)
                }
                is PairingViewModel.PairingState.Loading -> {
                    showLoading(true)
                    showError(null)
                }
                is PairingViewModel.PairingState.Success -> {
                    showLoading(false)
                    Toast.makeText(
                        requireContext(),
                        "Welcome, ${state.childName}!",
                        Toast.LENGTH_SHORT
                    ).show()
                    // Navigate to content list
                    findNavController().navigate(R.id.action_pairing_to_content)
                }
                is PairingViewModel.PairingState.Error -> {
                    showLoading(false)
                    showError(state.message)
                }
            }
        }
    }
    
    private fun showLoading(show: Boolean) {
        binding.progressBar.visibility = if (show) View.VISIBLE else View.GONE
        binding.btnPair.isEnabled = !show
        binding.etPairingCode.isEnabled = !show
    }
    
    private fun showError(message: String?) {
        binding.tvError.text = message ?: ""
        binding.tvError.visibility = if (message != null) View.VISIBLE else View.GONE
    }
    
    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
